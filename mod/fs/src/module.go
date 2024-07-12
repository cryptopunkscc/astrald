package fs

import (
	"context"
	"errors"
	"fmt"
	"github.com/cryptopunkscc/astrald/core/assets"
	"github.com/cryptopunkscc/astrald/events"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/mod/content"
	"github.com/cryptopunkscc/astrald/mod/fs"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/mod/sets"
	"github.com/cryptopunkscc/astrald/net"
	node2 "github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/object"
	"github.com/cryptopunkscc/astrald/sig"
	"gorm.io/gorm"
	"io"
	"os"
	"path/filepath"
	"strings"
)

var _ fs.Module = &Module{}
var defaultOpenOpts = &objects.OpenOpts{}

const workers = 1
const updatesLen = 1024

type Module struct {
	config Config
	node   node2.Node
	assets assets.Assets
	log    *log.Logger
	events events.Queue
	db     *gorm.DB
	ctx    context.Context

	objects objects.Module
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
func (mod *Module) Open(_ context.Context, objectID object.ID, opts *objects.OpenOpts) (objects.Reader, error) {
	if opts == nil {
		opts = defaultOpenOpts
	}

	if !opts.Zone.Is(net.ZoneDevice) {
		return nil, net.ErrZoneExcluded
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

	return nil, objects.ErrNotFound
}

func (mod *Module) Create(opts *objects.CreateOpts) (objects.Writer, error) {
	for _, dir := range mod.config.Store {
		r, err := mod.createObjectAt(dir, opts.Alloc)
		if err == nil {
			return r, err
		}
	}

	return nil, errors.New("no space available")
}

func (mod *Module) Purge(objectID object.ID, opts *objects.PurgeOpts) (count int, err error) {
	var id = objectID.String()

	for _, dir := range mod.config.Store {
		var path = filepath.Join(dir, id)

		if os.Remove(path) == nil {
			count++
		}
	}

	return
}

func (mod *Module) Find(opts *fs.FindOpts) (files []*fs.File) {
	if opts == nil {
		opts = &fs.FindOpts{}
	}

	var rows []*dbLocalFile

	var q = mod.db.Order("updated_at asc")

	if !opts.UpdatedAfter.IsZero() {
		q = q.Where("updated_at > ?", opts.UpdatedAfter)
	}

	err := q.Find(&rows).Error
	if err != nil {
		return
	}

	for _, row := range rows {
		files = append(files, &fs.File{
			Path:     row.Path,
			ObjectID: row.DataID,
			ModTime:  row.ModTime,
		})
	}

	return
}

func (mod *Module) Path(objectID object.ID) []string {
	return mod.path(objectID)
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
func (mod *Module) path(objectID object.ID) []string {
	var list []string

	var err = mod.db.
		Model(&dbLocalFile{}).
		Where("data_id = ?", objectID).
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

	if stat.ModTime().After(row.ModTime) {
		return errors.New("file modified")
	}

	return nil
}

// update updates the index entry for the path. path must be absolute.
func (mod *Module) update(path string) (object.ID, error) {
	if len(path) == 0 || path[0] != '/' {
		return object.ID{}, errors.New("invalid path")
	}

	stat, err := os.Stat(path)
	if err != nil {
		if err := mod.deletePath(path); err != nil {
			mod.log.Errorv(2, "deletePath: %v", err)
		}
		return object.ID{}, err
	}

	var row dbLocalFile
	err = mod.db.Where("path = ?", path).First(&row).Error
	if err == nil {
		var modtime = stat.ModTime()
		// if file hasn't changed, just return nil
		if !modtime.After(row.ModTime) {
			return row.DataID, nil
		}
	}

	objectID, err := object.ResolveFile(path)
	if err != nil {
		return object.ID{}, err
	}

	updated := &dbLocalFile{
		Path:    path,
		DataID:  objectID,
		ModTime: stat.ModTime(),
	}

	if row.Path == "" {
		err = mod.db.Create(updated).Error
		if err == nil {
			mod.events.Emit(fs.EventFileAdded{
				Path:     updated.Path,
				ObjectID: updated.DataID,
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

	if err == nil {
		mod.events.Emit(objects.EventDiscovered{
			ObjectID: updated.DataID,
			Zone:     net.ZoneDevice,
		})
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
		Path:     path,
		ObjectID: row.DataID,
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

func (mod *Module) createObjectAt(path string, alloc int) (objects.Writer, error) {
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
