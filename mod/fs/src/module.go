package fs

import (
	"context"
	"errors"
	"fmt"
	"github.com/cryptopunkscc/astrald/data"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/mod/content"
	"github.com/cryptopunkscc/astrald/mod/fs"
	"github.com/cryptopunkscc/astrald/mod/sets"
	"github.com/cryptopunkscc/astrald/mod/storage"
	"github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/node/assets"
	"github.com/cryptopunkscc/astrald/node/events"
	"github.com/cryptopunkscc/astrald/sig"
	"gorm.io/gorm"
	"io"
	"os"
	"path/filepath"
	"strings"
)

var _ fs.Module = &Module{}
var defaultOpenOpts = &storage.OpenOpts{}

const workers = 1
const updatesLen = 1024

type Module struct {
	config Config
	node   node.Node
	assets assets.Assets
	log    *log.Logger
	events events.Queue
	db     *gorm.DB
	ctx    context.Context

	storage storage.Module
	content content.Module
	sets    sets.Module

	watcher *Watcher
	updates chan sig.Task
}

func (mod *Module) Run(ctx context.Context) error {
	mod.ctx = ctx

	updatesDone := sig.Workers(ctx, mod.updates, workers)

	go mod.verifyIndex(ctx)

	<-ctx.Done()
	<-updatesDone

	return nil
}

// Open an object from the local filesystem
func (mod *Module) Open(objectID data.ID, opts *storage.OpenOpts) (storage.Reader, error) {
	if opts == nil {
		opts = defaultOpenOpts
	}

	paths := mod.path(objectID)
	for _, path := range paths {
		// check if the index for the path is valid
		if mod.validate(path) != nil {
			mod.enqueueUpdate(path)
			continue
		}

		f, err := os.Open(path)
		if err != nil {
			continue
		}

		n, err := f.Seek(int64(opts.Offset), io.SeekStart)
		if err != nil {
			f.Close()
			continue
		}

		if uint64(n) != opts.Offset {
			f.Close()
			continue
		}

		return &Reader{
			ReadSeekCloser: f,
			name:           path,
		}, nil
	}

	return nil, storage.ErrNotFound
}

func (mod *Module) Create(opts *storage.CreateOpts) (storage.Writer, error) {
	for _, dir := range mod.config.Store {
		r, err := mod.createObjectAt(dir, opts.Alloc)
		if err == nil {
			return r, err
		}
	}

	return nil, errors.New("no space available")
}

func (mod *Module) Find(ctx context.Context, query string, opts *content.FindOpts) (matches []content.Match, err error) {
	var rows []*dbLocalFile

	err = mod.db.
		Where("LOWER(PATH) like ?", "%"+strings.ToLower(query)+"%").
		Find(&rows).
		Error
	if err != nil {
		return
	}

	for _, row := range rows {
		matches = append(matches, content.Match{
			DataID: row.DataID,
			Score:  100,
			Exp:    "file path contains query",
		})
	}

	return
}

// Watch a directory tree for updates
func (mod *Module) Watch(path string) (added []string, err error) {
	mod.watcher.Add(path, false)

	entries, err := os.ReadDir(path)
	if err != nil {
		return
	}

	added = append(added, path)

	for _, entry := range entries {
		entryPath := filepath.Join(path, entry.Name())
		if entry.IsDir() {
			sub, _ := mod.Watch(entryPath)
			added = append(added, sub...)
		} else {
			mod.enqueueUpdate(entryPath)
		}
	}

	return
}

// path retruns a list of known paths to an object in the filesystem
func (mod *Module) path(dataID data.ID) []string {
	var list []string

	var err = mod.db.
		Model(&dbLocalFile{}).
		Where("data_id = ?", dataID).
		Select("path").
		Find(&list).
		Error

	if err != nil {
		mod.log.Error("path: database error: %v", err)
	}

	return list
}

// validate checks if the index is up-to-date for the given path
func (mod *Module) validate(path string) error {
	if len(path) == 0 || path[0] != '/' {
		return errors.New("invalid path")
	}

	stat, err := os.Stat(path)
	if err != nil {
		return fmt.Errorf("cannot access file: %v", err)
	}

	var row dbLocalFile
	err = mod.db.Where("path = ?", path).First(&row).Error
	if err != nil {
		return errors.New("not indexed")
	}

	if stat.ModTime().After(row.IndexedAt) {
		return errors.New("file modified")
	}

	return nil
}

// update updates the index entry for the path. path must be absolute.
func (mod *Module) update(path string) (data.ID, error) {
	if len(path) == 0 || path[0] != '/' {
		return data.ID{}, errors.New("invalid path")
	}

	stat, err := os.Stat(path)
	if err != nil {
		if err := mod.deletePath(path); err != nil {
			mod.log.Errorv(2, "deletePath: %v", err)
		}
		return data.ID{}, err
	}

	var row dbLocalFile
	err = mod.db.Where("path = ?", path).First(&row).Error
	if err == nil {
		var modtime = stat.ModTime()
		// if file hasn't changed, just return nil
		if !modtime.After(row.IndexedAt) {
			return row.DataID, nil
		}
	}

	dataID, err := data.ResolveFile(path)
	if err != nil {
		return data.ID{}, err
	}

	updated := &dbLocalFile{
		Path:      path,
		DataID:    dataID,
		IndexedAt: stat.ModTime(),
	}

	if row.Path == "" {
		err = mod.db.Create(updated).Error
		if err == nil {
			mod.events.Emit(fs.EventFileAdded{
				Path:   updated.Path,
				DataID: updated.DataID,
			})
		}
	} else {
		err = mod.db.
			Where("path = ?", path).
			Save(updated).
			Error
		if err == nil {
			mod.events.Emit(fs.EventFileChanged{
				Path:  updated.Path,
				OldID: row.DataID,
				NewID: updated.DataID,
			})
		}
	}

	return updated.DataID, err
}

func (mod *Module) deletePath(path string) error {
	var row *dbLocalFile

	err := mod.db.Where("path = ?", path).First(&row).Error
	if err != nil {
		return err
	}

	err = mod.db.Where("path = ?", path).Delete(&row).Error
	if err != nil {
		return err
	}

	mod.events.Emit(fs.EventFileRemoved{
		Path:   path,
		DataID: row.DataID,
	})

	return nil
}

// rename renames a file and updates the index without rescanning the file
func (mod *Module) rename(oldPath, newPath string) error {
	if len(oldPath) == 0 || oldPath[0] != '/' {
		return errors.New("invalid old path")
	}

	if len(newPath) == 0 || newPath[0] != '/' {
		return errors.New("invalid new path")
	}

	var row dbLocalFile
	err := mod.db.Where("path = ?", oldPath).First(&row).Error
	if err != nil {
		return errors.New("path not indexed")
	}

	if _, err := os.Stat(newPath); !os.IsNotExist(err) {
		return errors.New("new path already exists")
	}

	err = os.Rename(oldPath, newPath)
	if err != nil {
		return err
	}

	return mod.db.
		Model(&dbLocalFile{}).
		Where("path = ?", oldPath).
		Update("path", newPath).
		Error
}

func (mod *Module) onWriteDone(path string) {
	if mod.isPathIgnored(path) {
		return
	}

	mod.enqueueUpdate(path)
}

func (mod *Module) enqueueUpdate(path string) {
	mod.updates <- func(_ context.Context) {
		_, _ = mod.update(path)
	}
}

func (mod *Module) isPathIgnored(path string) bool {
	var filename = filepath.Base(path)

	if strings.HasPrefix(filename, tempFilePrefix) {
		return true
	}

	return false
}

func (mod *Module) createObjectAt(path string, alloc int) (storage.Writer, error) {
	usage, err := DiskUsage(path)
	if err != nil {
		return nil, err
	}

	if usage.Free < uint64(alloc) {
		return nil, errors.New("not enough space left")
	}

	w, err := NewWriter(mod, path)

	return w, err
}

func (mod *Module) verifyIndex(ctx context.Context) {
	var rows []*dbLocalFile

	err := mod.db.Find(&rows).Error
	if err != nil {
		mod.log.Error("error scanning index: %v", err)
	}
	for _, row := range rows {
		if mod.validate(row.Path) != nil {
			mod.log.Log("updating %v", row.Path)
			mod.enqueueUpdate(row.Path)
		}
		select {
		case <-ctx.Done():
			return
		default:
		}
	}

	mod.log.Log("done scanning index for changes")
}
