package fs

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/fs"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/sig"
)

var _ objects.Repository = &WatchRepository{}
var _ objects.AfterRemovedCallback = &WatchRepository{}

type WatchRepository struct {
	mod     *Module
	label   string
	root    string
	watcher *Watcher
}

func NewWatchRepository(mod *Module, root string, label string) (repo *WatchRepository, err error) {
	if !filepath.IsAbs(root) {
		return nil, fs.ErrNotAbsolute
	}

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

	repo.watcher.OnRenamed = repo.onChange
	repo.watcher.OnWriteDone = repo.onChange
	repo.watcher.OnRemoved = repo.onRemove

	repo.watcher.OnDirCreated = func(s string) {
		repo.watcher.Add(s, true)
	}

	repo.watcher.Add(root, true)

	// indexer will know to scan this root while init
	err = repo.mod.indexer.addRoot(root)
	if err != nil {
		return nil, err
	}

	return
}

var _ objects.Repository = &WatchRepository{}

func (repo *WatchRepository) Contains(ctx *astral.Context, objectID *astral.ObjectID) (bool, error) {
	return repo.mod.db.ObjectExists(repo.root, objectID)
}

func (repo *WatchRepository) onChange(path string) {
	repo.mod.indexer.invalidate(path)
}

func (repo *WatchRepository) onRemove(path string) {
	repo.mod.indexer.remove(path)
}

func (repo *WatchRepository) Scan(ctx *astral.Context, follow bool) (<-chan *astral.ObjectID, error) {
	ch := make(chan *astral.ObjectID)

	go func() {
		defer close(ch)
		defer func() {
			if follow {
				repo.mod.indexer.unsubscribe()
			}
		}()

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

		if follow {
			subscribe := sig.Subscribe(ctx, repo.mod.indexer.subscribe())
			for event := range subscribe {
				if pathUnderRoot(event.Path, repo.root) {
					select {
					case ch <- event.ObjectID:
					case <-ctx.Done():
						return
					}
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

func (repo *WatchRepository) AfterRemoved(name string) {
	if err := repo.watcher.Close(); err != nil {
		repo.mod.log.Error("%v watcher close error: %v", name, err)
	}

	if err := repo.mod.indexer.removeRoot(repo.root); err != nil {
		repo.mod.log.Error("%v indexer remove root error: %v", name, err)
	}
}
