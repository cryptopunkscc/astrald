package fs

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/objects"
)

type WatchRepository struct {
	mod     *Module
	label   string
	root    string
	watcher *Watcher

	scanCancel func()
	scanDone   <-chan error
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
		mod:   mod,
		label: label,
		root:  root,
	}
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

	// Register interest in this root
	repo.mod.fileIndexer.AcquireRoot(root)

	// Run initial scan as a cancelable job.
	repo.startScan(astral.NewContext(nil))

	return
}

var _ objects.Repository = &WatchRepository{}

func (repo *WatchRepository) Contains(ctx *astral.Context, objectID *astral.ObjectID) (bool, error) {
	return repo.mod.db.ObjectExists(repo.root, objectID)
}

func (repo *WatchRepository) onChange(path string) {
	repo.mod.fileIndexer.MarkDirty(path)
}

func (repo *WatchRepository) onRemove(path string) {
	repo.mod.fileIndexer.MarkDirty(path)
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

// startScan starts a background scan of repo.root.
// Cancels any previous scan before starting a new one.
func (repo *WatchRepository) startScan(parent *astral.Context) {
	// Cancel any previous scan and wait for it
	if repo.scanCancel != nil && repo.scanDone != nil {
		repo.scanCancel()
		<-repo.scanDone
	}

	// Start new scan
	ctx, cancel := parent.WithCancel()
	done := make(chan error, 1)

	repo.scanCancel = cancel
	repo.scanDone = done

	go func() {
		defer close(done)
		done <- repo.rescan(ctx)
	}()
}

func (repo *WatchRepository) rescan(ctx *astral.Context) error {
	return filepath.WalkDir(repo.root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Check context early for prompt cancellation.
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// Check if the entry is a regular file
		if !entry.Type().IsRegular() {
			return nil
		}

		err = repo.mod.checkIndexEntry(path)
		if err != nil {

			repo.mod.fileIndexer.MarkDirty(path)
		}

		return nil
	})
}

// Close stops background activity associated with the repository.
// It is safe to call multiple times.
func (repo *WatchRepository) Close() error {
	// Cancel scan and wait for it to actually finish
	if repo.scanCancel != nil && repo.scanDone != nil {
		repo.scanCancel()
		// Block until scan goroutine exits
		// This ensures no more MarkDirty calls after ReleaseRoot
		<-repo.scanDone
		// Clear to prevent double-wait on second Close()
		repo.scanCancel = nil
		repo.scanDone = nil
	}

	// Stop filesystem watcher
	if repo.watcher != nil {
		_ = repo.watcher.Close()
		repo.watcher = nil
	}

	// Release root interest (cleans up all file state automatically)
	repo.mod.fileIndexer.ReleaseRoot(repo.root)

	return nil
}
