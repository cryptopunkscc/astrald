package fs

import (
	"context"
	"errors"
	"fmt"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/log"
	"github.com/cryptopunkscc/astrald/core/assets"
	"github.com/cryptopunkscc/astrald/mod/fs"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/mod/shell"
	"github.com/cryptopunkscc/astrald/object"
	"github.com/cryptopunkscc/astrald/sig"
	"gorm.io/gorm"
	"os"
	"path/filepath"
	"strings"
)

var _ fs.Module = &Module{}
var defaultOpenOpts = &objects.OpenOpts{}

const workers = 1
const updatesLen = 4096

type Module struct {
	Deps
	config Config
	node   astral.Node
	assets assets.Assets
	log    *log.Logger
	db     *gorm.DB
	ctx    context.Context

	repos sig.Map[string, *Repository]

	watcher *Watcher
	updates chan sig.Task
	shares  sig.Map[string, *sig.Set[string]]

	ops shell.Scope
}

func (mod *Module) Run(ctx *astral.Context) error {
	mod.ctx = ctx

	for _, s := range mod.config.Shares {
		mod.shares.Set(s.Path, &sig.Set[string]{})
		share, _ := mod.shares.Get(s.Path)

		for _, name := range s.Allow {
			id, err := mod.Dir.ResolveIdentity(name)
			if err != nil {
				mod.log.Error("config: cannot resolve identity %v: %v", name, err)
			}
			share.Add(id.String())
		}

		mod.log.Infov(1, "%v shared with %v", s.Path, share.Clone())
	}

	updatesDone := sig.Workers(ctx, mod.updates, workers)

	for _, path := range mod.config.Watch {
		mod.Watch(path)
	}

	for _, path := range mod.config.Repos {
		mod.Watch(path)
	}

	go mod.verifyIndex(ctx)

	<-ctx.Done()
	<-updatesDone

	return nil
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

func (mod *Module) Scope() *shell.Scope {
	return &mod.ops
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

	objectID, err := resolveFileID(path)
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
			mod.Objects.Receive(&fs.EventFileAdded{
				Path:     astral.String16(updated.Path),
				ObjectID: updated.DataID,
			}, nil)
		}
	} else {
		err = mod.db.
			Where("path = ?", path).
			Save(updated).
			Error
		if err == nil {
			if !row.DataID.IsEqual(updated.DataID) {
				mod.Objects.Receive(&fs.EventFileChanged{
					Path:  astral.String16(updated.Path),
					OldID: row.DataID,
					NewID: updated.DataID,
				}, nil)
			}
		}
	}

	if err == nil {
		mod.Objects.Receive(&objects.EventDiscovered{
			ObjectID: updated.DataID,
			Zone:     astral.ZoneDevice,
		}, nil)
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

	mod.Objects.Receive(&fs.EventFileRemoved{
		Path:     astral.String16(path),
		ObjectID: row.DataID,
	}, nil)

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

func (mod *Module) String() string {
	return fs.ModuleName
}

func resolveFileID(path string) (object.ID, error) {
	file, err := os.Open(path)
	if err != nil {
		return object.ID{}, err
	}
	defer file.Close()

	fileID, err := object.Resolve(file)
	if err != nil {
		return object.ID{}, err
	}

	return fileID, nil
}
