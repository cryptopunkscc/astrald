package fs

import (
	"errors"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/sig"
)

type WatchRepository struct {
	mod      *Module
	label    string
	root     string
	watcher  *Watcher
	token    chan struct{}
	addQueue *sig.Queue[*astral.ObjectID]
}

func NewWatchRepository(mod *Module, root string, label string) (repo *WatchRepository, err error) {
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
		addQueue: &sig.Queue[*astral.ObjectID]{},
		token:    make(chan struct{}, 1),
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

	go repo.rescan(astral.NewContext(nil))

	return
}

var _ objects.Repository = &WatchRepository{}

func (repo *WatchRepository) Contains(ctx *astral.Context, objectID *astral.ObjectID) (bool, error) {
	return repo.mod.db.ObjectExists(repo.root, objectID)
}

func (repo *WatchRepository) onChange(path string) {
	<-repo.token // take a token

	objectID, err := repo.mod.update(path)
	if err == nil {
		repo.pushAdded(objectID)
	}

	repo.token <- struct{}{} // return a token
}

func (repo *WatchRepository) onRemove(path string) {
	repo.mod.update(path)
}

func (repo *WatchRepository) Scan(ctx *astral.Context, follow bool) (<-chan *astral.ObjectID, error) {
	ch := make(chan *astral.ObjectID)

	var subscribe <-chan *astral.ObjectID

	go func() {
		defer close(ch)

		if follow {
			subscribe = repo.addQueue.Subscribe(ctx)
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

func (repo *WatchRepository) rescan(ctx *astral.Context) error {
	filepath.WalkDir(repo.root, func(path string, entry os.DirEntry, err error) error {
		if err != nil {
			return err
		}

		// Check if the entry is a regular file
		if !entry.Type().IsRegular() {
			return nil
		}

		err = repo.mod.validate(path)
		if err != nil {
			repo.onChange(path)
		}

		return nil
	})

	return nil
}

func (repo *WatchRepository) pushAdded(id *astral.ObjectID) {
	repo.addQueue = repo.addQueue.Push(id)
}
