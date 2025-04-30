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
	root     string
	addQueue *sig.Queue[*object.ID]
}

var _ objects.Repository = &Repository{}

func NewRepository(mod *Module, name string, path string) *Repository {
	return &Repository{
		mod:      mod,
		name:     name,
		root:     path,
		addQueue: &sig.Queue[*object.ID]{},
	}
}

func (repo *Repository) Scan(ctx *astral.Context, follow bool) (<-chan *object.ID, error) {
	ch := make(chan *object.ID)

	var subsribe <-chan *object.ID

	go func() {
		defer close(ch)

		if follow {
			subsribe = repo.addQueue.Subscribe(ctx)
		}

		entries, err := os.ReadDir(repo.root)
		if err != nil {
			repo.mod.log.Error("cannot read dir %v: %v", repo.root, err)
			return
		}

		for _, entry := range entries {
			if !entry.Type().IsRegular() {
				continue
			}

			objectID, err := object.ParseID(entry.Name())
			if err != nil {
				continue
			}

			select {
			case ch <- objectID:
			case <-ctx.Done():
				return
			}

		}

		// handle subscription
		if subsribe != nil {
			for id := range subsribe {
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

func (repo *Repository) Read(ctx *astral.Context, objectID *object.ID, offset int64, limit int64) (objects.Reader, error) {
	if !ctx.Zone().Is(astral.ZoneDevice) {
		return nil, astral.ErrZoneExcluded
	}

	if limit == 0 {
		limit = int64(objectID.Size)
	}

	path := filepath.Join(repo.root, objectID.String())

	f, err := os.Open(path)
	if err != nil {
		return nil, objects.ErrNotFound
	}

	if offset != 0 {
		pos, err := f.Seek(offset, io.SeekStart)
		if err != nil {
			f.Close()
			return nil, objects.ErrNotFound
		}

		if pos != offset {
			f.Close()
			return nil, objects.ErrNotFound
		}
	}

	return NewReader(f, path, limit), nil

}

func (repo *Repository) Contains(ctx *astral.Context, objectID *object.ID) (bool, error) {
	path := filepath.Join(repo.root, objectID.String())

	// check if we have the file
	f, err := os.Stat(path)
	if err != nil {
		return false, nil
	}

	// and it's a regular file
	if f.Mode().IsRegular() {
		return true, nil
	}
	return false, nil
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

	w, err := NewWriter(repo, repo.root)

	return w, err
}

func (repo *Repository) Free(ctx *astral.Context) (int64, error) {
	usage, err := DiskUsage(repo.root)
	if err != nil {
		return -1, errors.ErrUnsupported
	}

	return int64(usage.Free), nil
}

func (repo *Repository) Delete(ctx *astral.Context, objectID *object.ID) error {
	path := filepath.Join(repo.root, objectID.String())
	return os.Remove(path)
}

func (repo *Repository) pushAdded(id *object.ID) {
	repo.addQueue = repo.addQueue.Push(id)
}
