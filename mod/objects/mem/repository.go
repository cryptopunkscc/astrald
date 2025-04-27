package mem

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/object"
	"github.com/cryptopunkscc/astrald/sig"
	"slices"
	"sync/atomic"
)

var _ objects.Repository = &Repository{}

const DefaultSize = 64 * 1024 * 1024 // 64MB

type Repository struct {
	objects sig.Map[string, []byte]
	mod     objects.Module
	used    atomic.Int64
	size    int64
	name    string
}

var _ objects.Repository = &Repository{}

func NewRepository(name string, size int64) *Repository {
	var repo = &Repository{name: name, size: size}

	if len(repo.name) == 0 {
		repo.name = "Memory"
	}
	if repo.size == 0 {
		repo.size = DefaultSize
	}
	return repo
}

func (repo *Repository) Label() string {
	return repo.name
}

func (repo *Repository) Create(ctx *astral.Context, opts *objects.CreateOpts) (objects.Writer, error) {
	if free, err := repo.Free(ctx); err == nil {
		if opts != nil && int64(opts.Alloc) > free {
			return nil, objects.ErrNoSpaceLeft
		}
	}

	return NewWriter(repo), nil
}

func (repo *Repository) Contains(ctx *astral.Context, objectID *object.ID) (bool, error) {
	return slices.Contains(repo.objects.Keys(), objectID.String()), nil
}

func (repo *Repository) Read(ctx *astral.Context, objectID *object.ID, offset int64, limit int64) (objects.Reader, error) {
	if !ctx.Zone().Is(astral.ZoneDevice) {
		return nil, astral.ErrZoneExcluded
	}

	s, e, err := getSliceBounds(objectID, offset, limit)
	if err != nil {
		return nil, err
	}

	bytes, found := repo.objects.Get(objectID.String())
	if !found {
		return nil, objects.ErrNotFound
	}

	return NewReader(bytes[s:e]), nil
}

func (repo *Repository) Scan(ctx *astral.Context) (<-chan *object.ID, error) {
	ch := make(chan *object.ID)

	go func() {
		defer close(ch)

		for _, s := range repo.objects.Keys() {
			id, err := object.ParseID(s)
			if err != nil {
				continue
			}
			select {
			case <-ctx.Done():
				return
			case ch <- id:
			}
		}
	}()

	return ch, nil
}

func (repo *Repository) Delete(ctx *astral.Context, objectID *object.ID) error {
	_, ok := repo.objects.Delete(objectID.String())
	if !ok {
		return objects.ErrNotFound
	}
	return nil
}

func (repo *Repository) Used() int64 {
	return repo.used.Load()
}

func (repo *Repository) Free(ctx *astral.Context) (int64, error) {
	return repo.size - repo.used.Load(), nil
}

func (repo *Repository) free() int64 {
	return repo.size - repo.used.Load()
}

func getSliceBounds(objectID *object.ID, offset int64, limit int64) (s int, e int, err error) {
	switch {
	case offset < 0 || offset > int64(objectID.Size):
		return 0, 0, objects.ErrOutOfBounds
	case limit < 0:
		return 0, 0, objects.ErrOutOfBounds
	case limit == 0:
		limit = int64(objectID.Size)
	}

	s = int(offset)
	e = min(s+int(limit), int(objectID.Size))

	return
}
