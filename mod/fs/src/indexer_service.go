package fs

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/data"
	"github.com/cryptopunkscc/astrald/mod/fs"
	"github.com/cryptopunkscc/astrald/mod/storage"
	"github.com/fsnotify/fsnotify"
	"io"
	"os"
	"path/filepath"
	"sync"
	"time"
)

const fileQueueSize = 1 << 18
const dirQueueSize = 1 << 14
const updateDelay = 3 * time.Second

type IndexService struct {
	*Module
	cond      *sync.Cond
	dirQueue  chan string
	fileQueue chan string

	watcher  *fsnotify.Watcher
	modified map[string]time.Time
	mu       sync.Mutex
}

func NewIndexService(mod *Module) *IndexService {
	return &IndexService{
		Module:    mod,
		cond:      sync.NewCond(&sync.Mutex{}),
		modified:  map[string]time.Time{},
		dirQueue:  make(chan string, dirQueueSize),
		fileQueue: make(chan string, fileQueueSize),
	}
}

func (srv *IndexService) Run(ctx context.Context) error {
	var err error

	srv.watcher, err = fsnotify.NewWatcher()
	if err != nil {
		srv.log.Error("filesystem watcher: %v", err)
	} else {
		go srv.watcherWorker(ctx)
		defer srv.watcher.Close()
	}

	go srv.verifyTree(ctx, "")
	go srv.dirWorker(ctx)
	go srv.fileWorker(ctx)

	<-ctx.Done()
	return nil
}

func (srv *IndexService) Read(id data.ID, opts *storage.ReadOpts) (storage.DataReader, error) {
	if opts == nil {
		opts = &storage.ReadOpts{}
	}

	files := srv.dbFindByID(id)

	for _, file := range files {
		info, err := os.Stat(file.Path)
		if err != nil {
			continue
		}

		if info.ModTime().After(file.IndexedAt) {
			continue
		}

		f, err := os.Open(file.Path)
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
			ReadCloser: f,
			name:       nameReadOnly,
		}, nil
	}

	return nil, storage.ErrNotFound
}

func (srv *IndexService) IndexSince(time time.Time) []storage.IndexEntry {
	var list []storage.IndexEntry

	var rows []*dbLocalFile
	var tx = srv.db.Where("indexed_at > ?", time).Find(&rows).Order("indexed_at")
	if tx.Error != nil {
		return nil
	}

	for _, row := range rows {
		dataID, err := data.Parse(row.DataID)
		if err != nil {
			continue
		}

		list = append(list, storage.IndexEntry{
			ID:        dataID,
			IndexedAt: row.IndexedAt,
		})
	}

	return list
}

func (srv *IndexService) Add(path string) error {
	info, err := os.Stat(path)
	if err != nil {
		return err
	}
	if info.IsDir() {
		return srv.enqueueDir(path)
	}
	if info.Mode().IsRegular() {
		return srv.enqueueFile(path)
	}
	return errors.New("invalid path")
}

func (srv *IndexService) enqueueDir(path string) error {
	select {
	case srv.dirQueue <- path:
		return nil
	default:
	}

	return errors.New("queue full")
}

func (srv *IndexService) enqueueFile(path string) error {
	select {
	case srv.fileQueue <- path:
		return nil
	default:
	}

	return errors.New("queue full")
}

func (srv *IndexService) fileWorker(ctx context.Context) {
	for {
		var path string

		select {
		case <-ctx.Done():
			return
		case path = <-srv.fileQueue:
		}

		row := srv.dbFindByPath(path)
		if row != nil {
			srv.verifyRow(ctx, row)
		} else {
			srv.indexFile(ctx, path)
		}
	}
}

func (srv *IndexService) verifyRow(_ context.Context, row *dbLocalFile) error {
	info, err := os.Stat(row.Path)

	// check if path isn't accessible or isn't a file
	if (err != nil) || !info.Mode().IsRegular() {
		return srv.unindexRow(row)
	}

	// check if file has been modified since it's been indexed
	if info.ModTime().Before(row.IndexedAt) {
		// file hasn't changed since it was indexed
		return nil
	}

	return srv.reindexRow(row)
}

func (srv *IndexService) indexFile(ctx context.Context, path string) error {
	var indexedAt = time.Now()

	fileID, err := data.ResolveFile(path)
	if err != nil {
		srv.log.Error("index %v: %v", path, err)
		return nil
	}

	var tx = srv.db.Create(&dbLocalFile{
		Path:      path,
		DataID:    fileID.String(),
		IndexedAt: indexedAt,
	})

	if tx.Error != nil {
		srv.log.Error("index %v: %v", path, err)
		return tx.Error
	}

	srv.log.Logv(2, "indexed %v (%v)", path, fileID)

	srv.events.Emit(storage.EventDataAdded{
		ID:        fileID,
		IndexedAt: indexedAt,
	})

	return nil
}

func (srv *IndexService) reindexRow(row *dbLocalFile) error {
	row.IndexedAt = time.Now()
	oldID, _ := data.Parse(row.DataID)
	fileID, err := data.ResolveFile(row.Path)
	if err != nil {
		return err
	}
	row.DataID = fileID.String()

	// update the database
	err = srv.db.Where("path = ?", row.Path).Save(row).Error
	if err != nil {
		srv.log.Error("reindex %v: %v", row.Path, err)
		return err
	}

	srv.log.Logv(2, "reindexed %v (%v)", row.Path, fileID)

	srv.events.Emit(storage.EventDataRemoved{
		ID: oldID,
	})

	srv.events.Emit(storage.EventDataAdded{
		ID:        fileID,
		IndexedAt: row.IndexedAt,
	})

	srv.events.Emit(fs.EventLocalFileChanged{
		Path:      row.Path,
		OldID:     oldID,
		NewID:     fileID,
		IndexedAt: row.IndexedAt,
	})

	return nil
}

func (srv *IndexService) unindexRow(row *dbLocalFile) error {
	var err = srv.dbDeleteByPath(row.Path)
	if err != nil {
		srv.log.Error("unindex %v: %v", row.Path, err)
		return err
	}

	srv.log.Logv(2, "unindexed %v", row.Path)

	dataID, _ := data.Parse(row.DataID)
	srv.events.Emit(storage.EventDataRemoved{
		ID: dataID,
	})

	return nil
}

func (srv *IndexService) dirWorker(ctx context.Context) {
	for {
		var path string

		select {
		case <-ctx.Done():
			return
		case path = <-srv.dirQueue:
		}

		entries, err := os.ReadDir(path)
		if err != nil {
			srv.log.Errorv(1, "error reading dir %v: %v", path, err)
			continue
		}

		srv.watchDir(path)

		for _, entry := range entries {
			var fullPath = filepath.Join(path, entry.Name())

			info, err := entry.Info()
			if err != nil {
				continue
			}

			if info.Mode().IsRegular() {
				srv.enqueueFile(fullPath)
			}
			if info.Mode().IsDir() {
				srv.enqueueDir(fullPath)
			}
		}
	}
}

func (srv *IndexService) watcherWorker(ctx context.Context) error {
	for {
		select {
		case event, ok := <-srv.watcher.Events:
			if !ok {
				return nil
			}

			switch {
			case event.Op.Has(fsnotify.Write):
				srv.onFileModified(ctx, event.Name)

			case event.Op.Has(fsnotify.Create):
				info, err := os.Stat(event.Name)
				if err != nil {
					continue
				}

				if info.Mode().IsRegular() {
					srv.onFileModified(ctx, event.Name)
					continue
				}

				if info.Mode().IsDir() {
					srv.enqueueDir(event.Name)
				}

			case event.Op.Has(fsnotify.Remove):
				row := srv.dbFindByPath(event.Name)
				if row != nil {
					srv.enqueueFile(event.Name)
				} else {
					srv.log.Logv(1, "unwatch %v", event.Name)
				}
			}

		case _, ok := <-srv.watcher.Errors:
			if !ok {
				return nil
			}
		}
	}
}

func (srv *IndexService) watchDir(path string) error {
	if srv.watcher == nil {
		return nil
	}

	var err = srv.watcher.Add(path)
	if err != nil {
		srv.log.Error("watch %v: %v", path, err)
	} else {
		srv.log.Logv(2, "watch %v", path)
	}

	return err
}

func (srv *IndexService) unwatchDir(path string) error {
	if srv.watcher == nil {
		return nil
	}

	var err = srv.watcher.Remove(path)
	if err != nil {
		srv.log.Error("unwatch %v: %v", path, err)
	} else {
		srv.log.Logv(1, "unwatch %v", path)
	}

	go srv.verifyTree(srv.ctx, path)

	return err
}

func (srv *IndexService) onFileModified(ctx context.Context, path string) {
	srv.mu.Lock()
	defer srv.mu.Unlock()

	if _, found := srv.modified[path]; !found {
		go srv.fileModifiedTimeout(ctx, path)
	}

	srv.modified[path] = time.Now().Add(updateDelay)
}

func (srv *IndexService) fileModifiedTimeout(ctx context.Context, path string) {
	for {
		srv.mu.Lock()
		deadline, found := srv.modified[path]

		if !found {
			srv.mu.Unlock()
			return
		}

		if time.Now().After(deadline) {
			delete(srv.modified, path)
			srv.mu.Unlock()
			srv.enqueueFile(path)
			return
		}

		srv.mu.Unlock()

		select {
		case <-ctx.Done():
			return
		case <-time.After(time.Until(deadline)):
		}
	}
}

func (srv *IndexService) verifyTree(ctx context.Context, path string) {
	var rows []*dbLocalFile

	var tx = srv.db.Where("path = ? or path like ?", path, path+"/%").Find(&rows)
	if tx.Error != nil {
		srv.log.Error("reverify %s: %v", path, tx.Error)
		return
	}

	for _, row := range rows {
		select {
		case <-ctx.Done():
			break
		case srv.fileQueue <- row.Path:
		}
	}
}
