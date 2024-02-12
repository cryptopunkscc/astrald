package fs

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/data"
	"github.com/cryptopunkscc/astrald/mod/fs"
	"github.com/cryptopunkscc/astrald/mod/sets"
	"github.com/cryptopunkscc/astrald/mod/storage"
	"github.com/cryptopunkscc/astrald/sig"
	"io"
	"os"
	"path/filepath"
	"strings"
	"syscall"
)

type ReadWriteService struct {
	*Module
	paths   sig.Set[string]
	watcher *Watcher
}

func NewReadWriteService(mod *Module) (*ReadWriteService, error) {
	var err error
	srv := &ReadWriteService{Module: mod}

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

func (srv *ReadWriteService) Run(ctx context.Context) error {
	<-ctx.Done()
	return nil
}

func (srv *ReadWriteService) onWriteDone(path string) {
	var filename = filepath.Base(path)
	if strings.HasPrefix(filename, tempFilePrefix) {
		return
	}

	srv.verify(path)
}

func (srv *ReadWriteService) onRemoved(path string) {
	dataID, err := data.Parse(filepath.Base(path))
	if err != nil {
		return
	}

	srv.rwSet.Remove(dataID)
}

func (srv *ReadWriteService) verify(path string) {
	dataID, err := data.Parse(filepath.Base(path))
	if err != nil {
		srv.rescan(path)
		return
	}

	scan, err := srv.rwSet.Scan(&sets.ScanOpts{DataID: dataID})
	if len(scan) == 0 {
		srv.rescan(path)
		return
	}

	stat, err := os.Stat(path)
	if err != nil {
		return
	}

	if stat.ModTime().After(scan[0].UpdatedAt) {
		srv.rescan(path)
	}
}

func (srv *ReadWriteService) rescan(path string) {
	dataID, err := data.ResolveFile(path)
	if err != nil {
		return
	}

	var newPath = filepath.Join(filepath.Dir(path), dataID.String())

	if newPath != path {
		os.Rename(path, newPath)
	}

	srv.rwSet.Add(dataID)
}

func (srv *ReadWriteService) Open(dataID data.ID, opts *storage.OpenOpts) (storage.Reader, error) {
	if opts == nil {
		opts = &storage.OpenOpts{}
	}

	scan, _ := srv.rwSet.Scan(&sets.ScanOpts{DataID: dataID})
	if len(scan) == 0 {
		return nil, storage.ErrNotFound
	}

	for _, dir := range srv.paths.Clone() {
		var path = filepath.Join(dir, dataID.String())

		r, err := srv.readPath(path, int(opts.Offset))
		if err == nil {
			return &Reader{ReadSeekCloser: r, name: fs.ReadWriteSetName}, err
		}
	}

	srv.rwSet.Remove(dataID)

	return nil, storage.ErrNotFound
}

func (srv *ReadWriteService) readPath(path string, offset int) (io.ReadSeekCloser, error) {
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

func (srv *ReadWriteService) Create(opts *storage.CreateOpts) (storage.Writer, error) {
	for _, dir := range srv.paths.Clone() {
		r, err := srv.createPath(dir, opts.Alloc)
		if err == nil {
			return r, err
		}
	}

	return nil, errors.New("no space available")
}

func (srv *ReadWriteService) createPath(path string, alloc int) (storage.Writer, error) {
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

func (srv *ReadWriteService) Purge(dataID data.ID, opts *storage.PurgeOpts) (n int, err error) {
	filename := dataID.String()
	var errs []error
	for _, dir := range srv.paths.Clone() {
		path := filepath.Join(dir, filename)
		stat, err := os.Stat(path)
		if err != nil {
			continue
		}
		if !stat.Mode().IsRegular() {
			continue
		}
		err = os.Remove(path)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		n++
	}

	return n, errors.Join(errs...)
}

func (srv *ReadWriteService) AddPath(path string) error {
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

func (srv *ReadWriteService) RemovePath(path string) error {
	srv.watcher.Remove(path, false)
	return srv.paths.Remove(path)
}

func (srv *ReadWriteService) Paths() []string {
	return srv.paths.Clone()
}

func (srv *ReadWriteService) Delete(dataID data.ID) error {
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

func (srv *ReadWriteService) deletePath(dir string, dataID data.ID) error {
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
