package fs

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/data"
	"github.com/cryptopunkscc/astrald/mod/fs"
	"github.com/cryptopunkscc/astrald/mod/storage"
	"github.com/cryptopunkscc/astrald/sig"
	"io"
	"os"
	"path/filepath"
	"strings"
	"syscall"
)

type StoreService struct {
	*Module
	paths   sig.Set[string]
	watcher *Watcher
}

func NewStoreService(mod *Module) (*StoreService, error) {
	var err error
	srv := &StoreService{Module: mod}

	srv.watcher, err = NewWatcher()
	if err != nil {
		return nil, err
	}

	srv.watcher.onWriteDone = srv.onWriteDone
	srv.watcher.OnRenamed = srv.onRemoved
	srv.watcher.OnRemoved = srv.onRemoved
	srv.watcher.OnChmod = srv.verify

	return srv, nil
}

func (srv *StoreService) Run(ctx context.Context) error {
	<-ctx.Done()
	return nil
}

func (srv *StoreService) onWriteDone(path string) {
	var filename = filepath.Base(path)
	if strings.HasPrefix(filename, tempFilePrefix) {
		return
	}

	srv.verify(path)
}

func (srv *StoreService) onRemoved(path string) {
	dataID, err := data.Parse(filepath.Base(path))
	if err != nil {
		return
	}

	srv.sets.RemoveFromSet(fs.ReadWriteSetName, dataID)
}

func (srv *StoreService) verify(path string) {
	dataID, err := data.Parse(filepath.Base(path))
	if err != nil {
		srv.rescan(path)
		return
	}

	entry, err := srv.sets.Member(fs.ReadWriteSetName, dataID)
	if err != nil {
		srv.rescan(path)
		return
	}

	stat, err := os.Stat(path)
	if err != nil {
		return
	}

	if stat.ModTime().After(entry.UpdatedAt) {
		srv.rescan(path)
	}
}

func (srv *StoreService) rescan(path string) {
	dataID, err := data.ResolveFile(path)
	if err != nil {
		return
	}

	var newPath = filepath.Join(filepath.Dir(path), dataID.String())

	if newPath != path {
		os.Rename(path, newPath)
	}

	srv.sets.AddToSet(fs.ReadWriteSetName, dataID)
}

func (srv *StoreService) Read(dataID data.ID, opts *storage.ReadOpts) (storage.DataReader, error) {
	if opts == nil {
		opts = &storage.ReadOpts{}
	}

	if m, _ := srv.sets.Member(fs.ReadWriteSetName, dataID); m == nil {
		return nil, storage.ErrNotFound
	}

	for _, dir := range srv.paths.Clone() {
		var path = filepath.Join(dir, dataID.String())

		r, err := srv.readPath(path, int(opts.Offset))
		if err == nil {
			return &Reader{ReadCloser: r, name: fs.ReadWriteSetName}, err
		}
	}

	srv.sets.RemoveFromSet(fs.ReadWriteSetName, dataID)

	return nil, storage.ErrNotFound
}

func (srv *StoreService) readPath(path string, offset int) (io.ReadCloser, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	if offset > 0 {
		r, err := f.Seek(int64(offset), io.SeekStart)
		if err != nil {
			f.Close()
			return nil, err
		}

		if int(r) != offset {
			return nil, errors.New("seek failed")
		}
	}

	return f, nil
}

func (srv *StoreService) Store(opts *storage.StoreOpts) (storage.DataWriter, error) {
	for _, dir := range srv.paths.Clone() {
		r, err := srv.storePath(dir, opts.Alloc)
		if err == nil {
			return r, err
		}
	}

	return nil, errors.New("no space available")
}

func (srv *StoreService) storePath(path string, alloc int) (storage.DataWriter, error) {
	usage, err := DiskUsage(path)
	if err != nil {
		return nil, err
	}

	if usage.Free < uint64(alloc) {
		return nil, errors.New("not enough free space")
	}

	w, err := NewFileWriter(srv, path)

	return w, err
}

func (srv *StoreService) AddPath(path string) error {
	srv.watcher.Add(path, false)
	err := srv.paths.Add(path)
	if err != nil {
		return err
	}

	entries, err := os.ReadDir(path)
	if err != nil {
		return nil
	}

	for _, entry := range entries {
		srv.verify(filepath.Join(path, entry.Name()))
	}

	return nil
}

func (srv *StoreService) RemovePath(path string) error {
	srv.watcher.Remove(path, false)
	return srv.paths.Remove(path)
}

func (srv *StoreService) Paths() []string {
	return srv.paths.Clone()
}

func (srv *StoreService) Delete(dataID data.ID) error {
	var deleted bool

	for _, dir := range srv.paths.Clone() {
		err := srv.deletePath(dir, dataID)
		if err == nil {
			srv.events.Emit(fs.EventFileRemoved{
				DataID: dataID,
				Path:   dir,
			})
			deleted = true
		}
	}

	if deleted {
		return nil
	}

	return errors.New("not found")
}

func (srv *StoreService) deletePath(dir string, dataID data.ID) error {
	path := filepath.Join(dir, dataID.String())

	info, err := os.Stat(path)
	if err != nil {
		return storage.ErrNotFound
	}

	if !info.Mode().IsRegular() {
		return storage.ErrNotFound
	}

	return os.Remove(path)
}

func DiskUsage(path string) (usage *DiskUsageInfo, err error) {
	var fs syscall.Statfs_t
	err = syscall.Statfs(path, &fs)
	if err != nil {
		return nil, err
	}

	return &DiskUsageInfo{
		Total:     fs.Blocks * uint64(fs.Bsize),
		Free:      fs.Bfree * uint64(fs.Bsize),
		Available: fs.Bavail * uint64(fs.Bsize),
	}, nil
}

// DiskUsageInfo represents disk usage information
type DiskUsageInfo struct {
	Total     uint64
	Free      uint64
	Available uint64
}
