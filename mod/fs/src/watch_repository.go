package fs

import (
	"errors"
	"fmt"
	"io"
	"os"
	"sync"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/objects"
)

type WatchRepository struct {
	mod     *Module
	label   string
	root    string
	watcher *Watcher
	token   chan struct{}

	scan ScanHandle

	acquiredMu sync.Mutex
	acquired   map[string]struct{}
}

func NewWatchRepository(mod *Module, label string, root string) (repo *WatchRepository, err error) {
	stat, err := os.Stat(root)
	switch {
	case err != nil:
		return nil, err
	case !stat.IsDir():
		return nil, fmt.Errorf("path %v is not a directory", root)
	}

	repo = &WatchRepository{
		mod:      mod,
		label:    label,
		root:     root,
		token:    make(chan struct{}, 1),
		acquired: make(map[string]struct{}),
	}

	repo.token <- struct{}{}
	repo.watcher, err = NewWatcher()
	if err != nil {
		return nil, err
	}

	repo.watcher.OnWriteDone = repo.onChange
	repo.watcher.OnRemoved = repo.onRemove
	repo.watcher.OnRenamed = repo.onRemove
	repo.watcher.OnDirCreated = func(s string) {
		repo.watcher.Add(s, true)
	}

	repo.watcher.Add(root, true)

	// Run initial scan as a cancelable job.
	repo.scan = repo.StartScan(astral.NewContext(nil))

	return
}

var _ objects.Repository = &WatchRepository{}

func (repo *WatchRepository) Contains(ctx *astral.Context, objectID *astral.ObjectID) (bool, error) {
	return repo.mod.db.ObjectExists(repo.root, objectID)
}

// acquire registers interest in a path for this repository.
// Idempotent: calling multiple times for the same path is safe.
func (repo *WatchRepository) acquire(path string) {
	repo.acquiredMu.Lock()
	defer repo.acquiredMu.Unlock()

	if _, ok := repo.acquired[path]; ok {
		return
	}
	repo.acquired[path] = struct{}{}
	repo.mod.fileIndexer.Acquire(path)
}

func (repo *WatchRepository) onChange(path string) {
	<-repo.token // take a token
	defer func() { repo.token <- struct{}{} }()

	repo.acquire(path)
	repo.mod.fileIndexer.MarkDirty(path)
}

func (repo *WatchRepository) onRemove(path string) {
	repo.acquiredMu.Lock()
	_, alreadyAcquired := repo.acquired[path]
	repo.acquiredMu.Unlock()

	if alreadyAcquired {
		// Path was already acquired, just mark dirty
		repo.mod.fileIndexer.MarkDirty(path)
	} else {
		// Temporarily acquire, mark dirty, then release
		// This ensures updateDbIndex runs to clean up the DB row,
		// but we don't retain long-term interest for a removed path
		repo.mod.fileIndexer.Acquire(path)
		repo.mod.fileIndexer.MarkDirty(path)
		repo.mod.fileIndexer.Release(path)
	}
}

func (repo *WatchRepository) Scan(ctx *astral.Context, follow bool) (<-chan *astral.ObjectID, error) {
	ch := make(chan *astral.ObjectID)

	var subscribe <-chan *astral.ObjectID

	go func() {
		defer close(ch)

		if follow {
			subscribe = repo.mod.fileIndexer.Subscribe(ctx)
		}

		ids, err := repo.mod.db.UniqueObjectIDs(repo.root)
		if err != nil {
			repo.mod.log.Error("db error: %v", err)
			return
		}

		for _, id := range ids {
			select {
			case ch <- id:
			case <-ctx.Done():
				return
			}
		}

		// handle subscription
		if subscribe != nil {
			for id := range subscribe {
				// Filter to this repository root to avoid leaking updates from other roots.
				ok, err := repo.mod.db.ObjectExists(repo.root, id)
				if err != nil || !ok {
					continue
				}
				select {
				case ch <- id:
				case <-ctx.Done():
					return
				}
			}
		}
	}()

	return ch, nil
}

func (repo *WatchRepository) Read(ctx *astral.Context, objectID *astral.ObjectID, offset int64, limit int64) (io.ReadCloser, error) {
	rows, err := repo.mod.db.FindObject(repo.root, objectID)
	if err != nil {
		return nil, err
	}
	if len(rows) == 0 {
		return nil, objects.ErrNotFound
	}
	if limit == 0 {
		limit = int64(objectID.Size)
	}

	for _, row := range rows {
		f, err := os.Open(row.Path)
		if err != nil {
			continue
		}

		if offset != 0 {
			pos, err := f.Seek(offset, io.SeekStart)
			if err != nil || pos != offset {
				f.Close()
				continue
			}
		}

		return NewReader(f, row.Path, limit), nil
	}

	return nil, objects.ErrNotFound
}

func (repo *WatchRepository) Create(ctx *astral.Context, opts *objects.CreateOpts) (objects.Writer, error) {
	return nil, errors.ErrUnsupported
}

func (repo *WatchRepository) Label() string {
	return repo.label
}

func (repo *WatchRepository) Delete(ctx *astral.Context, objectID *astral.ObjectID) error {
	return errors.ErrUnsupported
}

func (repo *WatchRepository) Free(ctx *astral.Context) (int64, error) {
	return 0, nil
}

func (repo *WatchRepository) String() string {
	return repo.label
}

// Close stops background activity associated with the repository.
// It is safe to call multiple times.
func (repo *WatchRepository) Close() error {
	if repo.scan.Cancel != nil {
		repo.scan.Cancel()
	}
	if repo.watcher != nil {
		_ = repo.watcher.Close()
	}

	// Release all acquired paths
	repo.acquiredMu.Lock()
	for path := range repo.acquired {
		repo.mod.fileIndexer.Release(path)
	}
	repo.acquired = make(map[string]struct{})
	repo.acquiredMu.Unlock()

	return nil
}
