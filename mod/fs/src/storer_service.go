package fs

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/data"
	"github.com/cryptopunkscc/astrald/mod/storage"
	"github.com/cryptopunkscc/astrald/sig"
	"io"
	"os"
	"path/filepath"
	"slices"
	"syscall"
	"time"
)

type StoreService struct {
	*Module
	paths sig.Set[string]
}

func NewStoreService(mod *Module) *StoreService {
	return &StoreService{Module: mod}
}

func (srv *StoreService) Run(ctx context.Context) error {
	<-ctx.Done()
	return nil
}

func (srv *StoreService) Read(dataID data.ID, opts *storage.ReadOpts) (storage.DataReader, error) {
	if opts == nil {
		opts = &storage.ReadOpts{}
	}

	for _, dir := range srv.paths.Clone() {
		var path = filepath.Join(dir, dataID.String())

		r, err := srv.readPath(path, int(opts.Offset))
		if err == nil {
			return r, err
		}
	}

	return nil, storage.ErrNotFound
}

func (srv *StoreService) readPath(path string, offset int) (io.ReadCloser, error) {
	f, err := os.Open(path)
	if err != nil {
		return nil, err
	}

	if offset > 0 {
		r, err := f.Seek(io.SeekStart, offset)
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

func (srv *StoreService) Store(alloc int) (storage.DataWriter, error) {
	for _, dir := range srv.paths.Clone() {
		r, err := srv.storePath(dir, alloc)
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
	return srv.paths.Add(path)
}

func (srv *StoreService) RemovePath(path string) error {
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
			deleted = true
		}
	}

	if deleted {
		srv.events.Emit(storage.EventDataRemoved{
			ID: dataID,
		})
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

	os.Remove(path)

	return nil
}
func (srv *StoreService) IndexSince(since time.Time) []storage.DataInfo {
	var list []storage.DataInfo

	for _, dir := range srv.paths.Clone() {
		list = append(list, srv.readDirSince(dir, since)...)
	}

	slices.SortFunc(list, func(a, b storage.DataInfo) int {
		return a.IndexedAt.Compare(b.IndexedAt)
	})

	return list
}

func (srv *StoreService) readDirSince(dir string, since time.Time) []storage.DataInfo {
	var list []storage.DataInfo

	entries, err := os.ReadDir(dir)
	if err != nil {
		return nil
	}

	for _, entry := range entries {
		if !entry.Type().IsRegular() {
			continue
		}

		dataID, err := data.Parse(entry.Name())
		if err != nil {
			continue
		}

		info, err := entry.Info()
		if err != nil {
			continue
		}

		if info.ModTime().After(since) {
			list = append(list, storage.DataInfo{
				ID:        dataID,
				IndexedAt: info.ModTime(),
			})
		}
	}

	return list
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
