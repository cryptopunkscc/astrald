package fs

import (
	"errors"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/object"
	"github.com/cryptopunkscc/astrald/sig"
	"io"
	"os"
	"path/filepath"
)

type Repository struct {
	mod      *Module
	name     string
	path     string
	addQueue *sig.Queue[*object.ID]
}

var _ objects.Repository = &Repository{}

func NewRepository(mod *Module, name string, path string) *Repository {
	return &Repository{
		mod:      mod,
		name:     name,
		path:     path,
		addQueue: &sig.Queue[*object.ID]{},
	}
}

func (repo *Repository) Scan(ctx *astral.Context, follow bool) (<-chan *object.ID, error) {
	ch := make(chan *object.ID)

	var s <-chan *object.ID

	go func() {
		defer close(ch)

		if follow {
			s = repo.addQueue.Subscribe(ctx)
		}

		ids, err := repo.mod.db.UniqueObjectIDs(repo.path)
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

		if s != nil {
			for i := range s {
				ch <- i
			}
		}
	}()

	return ch, nil
}

func (repo *Repository) Read(ctx *astral.Context, objectID *object.ID, offset int64, limit int64) (objects.Reader, error) {
	if !ctx.Zone().Is(astral.ZoneDevice) {
		return nil, astral.ErrZoneExcluded
	}

	if limit == 0 {
		limit = int64(objectID.Size)
	}

	paths := repo.mod.localPaths(objectID)
	for _, path := range paths {
		// check if the index for the path is valid
		err := repo.mod.validate(path)
		if err != nil {
			repo.mod.enqueueUpdate(path) //TODO: immediade update & retry?
			continue
		}

		f, err := os.Open(path)
		if err != nil {
			continue
		}

		pos, err := f.Seek(offset, io.SeekStart)
		if err != nil {
			f.Close()
			continue
		}

		if pos != offset {
			f.Close()
			continue
		}

		return NewReader(f, path, limit), nil
	}

	return nil, objects.ErrNotFound
}

func (repo *Repository) Contains(ctx *astral.Context, objectID *object.ID) (bool, error) {
	return repo.mod.db.ObjectExists(objectID)
}

func (repo *Repository) Label() string {
	return repo.name
}

func (repo *Repository) Create(ctx *astral.Context, opts *objects.CreateOpts) (objects.Writer, error) {
	if free, err := repo.Free(nil); err == nil {
		if opts != nil && free < int64(opts.Alloc) {
			return nil, objects.ErrNoSpaceLeft
		}
	}

	w, err := NewWriter(repo, repo.path)

	return w, err
}

func (repo *Repository) Free(ctx *astral.Context) (int64, error) {
	usage, err := DiskUsage(repo.path)
	if err != nil {
		return -1, errors.ErrUnsupported
	}

	return int64(usage.Free), nil
}

func (repo *Repository) Delete(ctx *astral.Context, objectID *object.ID) error {
	path := filepath.Join(repo.path, objectID.String())
	return os.Remove(path)
}

func (repo *Repository) pushAdded(id *object.ID) {
	repo.addQueue = repo.addQueue.Push(id)
}
