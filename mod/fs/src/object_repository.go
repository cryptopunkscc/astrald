package fs

import (
	"errors"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/object"
	"io"
	"os"
	"path/filepath"
)

type Repository struct {
	mod  *Module
	name string
	path string
}

var _ objects.Repository = &Repository{}

func NewRepository(mod *Module, name string, path string) *Repository {
	return &Repository{mod: mod, name: name, path: path}
}

func (r Repository) Scan(_ *astral.Context) (<-chan *object.ID, error) {
	ch := make(chan *object.ID)

	go func() {
		defer close(ch)

		var ids []*object.ID

		tx := r.mod.db.
			Model(&dbLocalFile{}).
			Select("data_id").
			Where("path like ?", r.path+"%").
			Find(&ids)

		if tx.Error != nil {
			return
		}

		for _, id := range ids {
			ch <- id
		}
	}()

	return ch, nil
}

func (r Repository) Read(ctx *astral.Context, objectID *object.ID, offset int64, limit int64) (objects.Reader, error) {
	if !ctx.Zone().Is(astral.ZoneDevice) {
		return nil, astral.ErrZoneExcluded
	}

	if limit == 0 {
		limit = int64(objectID.Size)
	}

	paths := r.mod.path(objectID)
	for _, path := range paths {
		// check if the index for the path is valid
		err := r.mod.validate(path)
		if err != nil {
			r.mod.enqueueUpdate(path) //TODO: immediade update & retry?
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

func (r Repository) Contains(ctx *astral.Context, objectID *object.ID) (bool, error) {
	return r.mod.db.ObjectExists(objectID)
}

func (r Repository) Label() string {
	return r.name
}

func (r Repository) Create(ctx *astral.Context, opts *objects.CreateOpts) (objects.Writer, error) {
	if free, err := r.Free(nil); err == nil {
		if opts != nil && free < int64(opts.Alloc) {
			return nil, objects.ErrNoSpaceLeft
		}
	}

	w, err := NewWriter(r.mod, r.path)

	return w, err
}

func (r Repository) Free(ctx *astral.Context) (int64, error) {
	usage, err := DiskUsage(r.path)
	if err != nil {
		return -1, errors.ErrUnsupported
	}

	return int64(usage.Free), nil
}

func (r Repository) Delete(ctx *astral.Context, objectID *object.ID) error {
	path := filepath.Join(r.path, objectID.String())
	return os.Remove(path)
}
