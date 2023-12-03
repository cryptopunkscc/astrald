package storage

import (
	"context"
	"errors"
	"fmt"
	"github.com/cryptopunkscc/astrald/data"
	storage "github.com/cryptopunkscc/astrald/mod/storage/api"
	"github.com/fsnotify/fsnotify"
	"io"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"sync/atomic"
	"time"
)

// updateDelay - wait this long after latest file change before reindexing it
const updateDelay = 3 * time.Second

var _ storage.LocalFiles = &LocalFiles{}

type LocalFiles struct {
	*Module
	root    string
	watcher *fsnotify.Watcher
	running atomic.Bool

	pendingDirs []string
	dirs        map[string]any
	dirsMu      sync.Mutex

	mod map[string]time.Time
	mu  sync.Mutex
}

type LocalFile struct {
	Path string
	ID   data.ID
}

type dbFile struct {
	Path      string `gorm:"primaryKey,index"`
	DataID    string `gorm:"index"`
	IndexedAt time.Time
}

func (dbFile) TableName() string { return "files" }

func NewLocalFiles(module *Module) *LocalFiles {
	return &LocalFiles{
		Module: module,
		mod:    map[string]time.Time{},
		dirs:   map[string]any{},
	}
}

func (source *LocalFiles) AddDir(ctx context.Context, path string) error {
	source.dirsMu.Lock()
	defer source.dirsMu.Unlock()

	if !source.running.Load() {
		source.pendingDirs = append(source.pendingDirs, path)
		return nil
	}

	select {
	case <-ctx.Done():
		return ctx.Err()
	default:
	}

	if _, found := source.dirs[path]; found {
		return errors.New("already added")
	}

	info, err := os.Stat(path)
	if err != nil {
		fmt.Println(path, err)
		return err
	}

	if !info.IsDir() {
		return errors.New("not a directory")
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		return err
	}

	if source.watcher != nil {
		err = source.watcher.Add(path)
		if err != nil {
			return err
		}
	}

	source.dirs[path] = nil

	source.log.Logv(2, "watching %v", path)

	for _, entry := range entries {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		var entryPath = filepath.Join(path, entry.Name())

		if entry.IsDir() {
			go source.AddDir(ctx, entryPath)
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}
		if !info.Mode().IsRegular() {
			continue
		}

		source.updateFile(ctx, entryPath)
	}

	return nil
}

func (source *LocalFiles) RemoveDir(ctx context.Context, path string) error {
	source.dirsMu.Lock()
	defer source.dirsMu.Unlock()

	var found bool

	for d := range source.dirs {
		if d == path || strings.HasPrefix(d, path+"/") {
			found = true

			source.log.Logv(2, "stopped watching %v", d)

			if source.watcher != nil {
				source.watcher.Remove(d)
			}

			delete(source.dirs, d)
		}
	}

	source.invalidatePath(ctx, path)

	if !found {
		return storage.ErrNotFound
	}

	return nil
}

func (source *LocalFiles) DataSince(time time.Time) []storage.DataInfo {
	var list []storage.DataInfo
	var rows []*dbFile

	tx := source.db.Where("indexed_at > ?", time).Find(&rows).Order("indexed_at")

	if tx.Error != nil {
		return nil
	}

	for _, row := range rows {
		dataID, err := data.Parse(row.DataID)
		if err != nil {
			continue
		}

		list = append(list, storage.DataInfo{
			ID:        dataID,
			IndexedAt: row.IndexedAt,
		})
	}

	return list
}

func (source *LocalFiles) scanDir(ctx context.Context, path string) error {
	entries, err := os.ReadDir(path)
	if err != nil {
		return err
	}

	for _, entry := range entries {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		var entryPath = filepath.Join(path, entry.Name())

		info, err := entry.Info()
		if err == nil && info.Mode().IsRegular() {
			source.updateFile(ctx, entryPath)
		}
	}

	return nil
}

func (source *LocalFiles) Read(id data.ID, offset int, length int) (io.ReadCloser, error) {
	files := source.findByDataID(id.String())

	if len(files) == 0 {
		return nil, storage.ErrNotFound
	}

	for _, file := range files {
		info, err := os.Stat(file.Path)
		if err != nil {
			continue
		}
		if info.ModTime().After(file.IndexedAt) {
			continue
		}

		f, err := os.Open(file.Path)
		if err == nil {
			f.Seek(int64(offset), io.SeekStart)
			return f, nil
		}
	}

	return nil, storage.ErrNotFound
}

func (source *LocalFiles) Run(ctx context.Context) error {
	if !source.running.CompareAndSwap(false, true) {
		return errors.New("already running")
	}
	defer source.running.Store(false)

	var err error

	source.watcher, err = fsnotify.NewWatcher()
	if err != nil {
		source.log.Error("cannot watch static files: %v", err)
	} else {
		defer source.watcher.Close()
		go source.handleWatchEvents(ctx)

	}

	// scan for changes in indexed files
	source.invalidatePath(ctx, "/")

	source.dirsMu.Lock()
	var dirs = source.pendingDirs
	source.pendingDirs = nil
	source.dirsMu.Unlock()

	for _, d := range dirs {
		source.AddDir(ctx, d)
	}

	<-ctx.Done()
	return nil
}

func (source *LocalFiles) Rescan(ctx context.Context) {
	source.invalidatePath(ctx, "/")

	source.dirsMu.Lock()
	var dirs []string
	for d := range source.dirs {
		dirs = append(dirs, d)
	}
	source.dirsMu.Unlock()

	for _, d := range dirs {
		select {
		case <-ctx.Done():
			return
		default:
		}

		source.scanDir(ctx, d)
	}
}

func (source *LocalFiles) touch(ctx context.Context, path string) {
	source.mu.Lock()
	defer source.mu.Unlock()

	if _, found := source.mod[path]; !found {
		go source.modTimeout(ctx, path)
	}

	source.mod[path] = time.Now().Add(updateDelay)
}

func (source *LocalFiles) modTimeout(ctx context.Context, path string) {
	for {
		source.mu.Lock()
		deadline, found := source.mod[path]

		if !found {
			source.mu.Unlock()
			return
		}

		if time.Now().After(deadline) {
			delete(source.mod, path)
			source.mu.Unlock()
			source.updateFile(ctx, path)
			return
		}

		source.mu.Unlock()

		select {
		case <-ctx.Done():
			return
		case <-time.After(time.Until(deadline)):
		}
	}
}

func (source *LocalFiles) handleWatchEvents(ctx context.Context) error {
	for {
		select {
		case event, ok := <-source.watcher.Events:
			if !ok {
				return nil
			}

			switch {
			case event.Op.Has(fsnotify.Write):
				source.touch(ctx, event.Name)

			case event.Op.Has(fsnotify.Create):
				info, err := os.Stat(event.Name)
				if err != nil {
					continue
				}

				if info.Mode().IsRegular() {
					source.touch(ctx, event.Name)
					continue
				}

				if info.Mode().IsDir() {
					source.AddDir(ctx, event.Name)
				}

			case event.Op.Has(fsnotify.Remove):
				source.touch(ctx, event.Name)
				source.RemoveDir(ctx, event.Name)
			}

		case _, ok := <-source.watcher.Errors:
			if !ok {
				return nil
			}
		}
	}
}

func (source *LocalFiles) invalidatePath(ctx context.Context, path string) error {
	var rows []*dbFile

	err := source.db.Where("path like ?", path+"%").Find(&rows).Error
	if err != nil {
		return err
	}

	for _, row := range rows {
		select {
		case <-ctx.Done():
			break
		default:
		}

		source.updateRow(ctx, row)
	}

	return nil
}

func (source *LocalFiles) updateFile(ctx context.Context, path string) error {
	var row = source.findByPath(path)

	if row != nil {
		return source.updateRow(ctx, row)
	}

	var indexedAt = time.Now()

	fileID, err := source.resolveID(path)
	if err != nil {
		return nil
	}

	err = source.db.Create(&dbFile{
		Path:      path,
		DataID:    fileID.String(),
		IndexedAt: indexedAt,
	}).Error

	if err != nil {
		source.log.Error("add %v error: %v", path, err)
	} else {
		source.log.Logv(2, "added %v as %v", path, fileID)

		source.events.Emit(storage.EventLocalFileAdded{
			Path: path,
			ID:   fileID,
		})
	}

	return nil

}

func (source *LocalFiles) updateRow(_ context.Context, row *dbFile) error {
	info, err := os.Stat(row.Path)

	if (err != nil) || !info.Mode().IsRegular() {
		// file no longer exists
		if err := source.deleteByPath(row.Path); err != nil {
			source.log.Error("remove %v error: %v", row.Path, err)
		} else {
			source.log.Logv(2, "removed %v", row.Path)

			dataID, _ := data.Parse(row.DataID)

			source.events.Emit(storage.EventLocalFileRemoved{
				Path: row.Path,
				ID:   dataID,
			})
		}

		return nil
	}

	if info.ModTime().Before(row.IndexedAt) {
		// file hasn't changed since it was indexed
		return nil
	}

	row.IndexedAt = time.Now()
	oldID, _ := data.Parse(row.DataID)
	fileID, err := source.resolveID(row.Path)
	if err != nil {
		return nil
	}
	row.DataID = fileID.String()

	if err := source.db.Where("path = ?", row.Path).Save(row).Error; err != nil {
		source.log.Error("update %v error: %v", row.Path, err)
	} else {
		source.log.Logv(2, "updated %v", row.Path)

		source.events.Emit(storage.EventLocalFileChanged{
			Path:  row.Path,
			OldID: oldID,
			NewID: fileID,
		})
	}

	return nil
}

func (source *LocalFiles) deleteByPath(path string) error {
	return source.db.Where("path = ?", path).Delete(&dbFile{}).Error
}

func (source *LocalFiles) resolveID(path string) (data.ID, error) {
	file, err := os.Open(path)
	if err != nil {
		return data.ID{}, err
	}
	defer file.Close()

	fileID, err := data.ResolveAll(file)
	if err != nil {
		return data.ID{}, err
	}

	return fileID, nil
}

func (source *LocalFiles) findByPath(path string) *dbFile {
	var file dbFile
	tx := source.db.First(&file, "path = ?", path)

	if tx.Error != nil {
		return nil
	}

	return &file
}

func (source *LocalFiles) findByPrefix(path string) []*LocalFile {
	var files []*LocalFile

	var rows []*dbFile
	tx := source.db.Where("path like ?", path+"%").Order("path").Find(&rows)

	if tx.Error != nil {
		return nil
	}

	for _, row := range rows {
		dataID, err := data.Parse(row.DataID)
		if err != nil {
			continue
		}

		files = append(files, &LocalFile{
			Path: row.Path,
			ID:   dataID,
		})
	}

	return files
}

func (source *LocalFiles) findByDataID(dataID string) []*dbFile {
	var files []*dbFile
	tx := source.db.Where("data_id = ?", dataID).Find(&files)
	if tx.Error != nil {
		return nil
	}
	return files
}
