package mem

import (
	"context"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/object"
	"github.com/cryptopunkscc/astrald/sig"
	"sync/atomic"
)

var _ objects.Opener = &Store{}
var _ objects.Creator = &Store{}
var _ objects.Purger = &Store{}

const DefaultSize = 64 * 1024 * 1024 // 64MB

type Store struct {
	objects sig.Map[string, []byte]
	mod     objects.Module
	used    atomic.Int64
	size    int64
}

func NewMemStore(mod objects.Module, size int64) *Store {
	var mem = &Store{mod: mod, size: DefaultSize}
	if size != 0 {
		mem.size = size
	}
	return mem
}

func (mem *Store) Open(_ context.Context, objectID object.ID, opts *objects.OpenOpts) (objects.Reader, error) {
	if !opts.Zone.Is(astral.ZoneDevice) {
		return nil, astral.ErrZoneExcluded
	}

	bytes, found := mem.objects.Get(objectID.String())
	if !found {
		return nil, objects.ErrNotFound
	}

	return NewMemDataReader(bytes), nil
}

func (mem *Store) Purge(objectID object.ID, opts *objects.PurgeOpts) (int, error) {
	_, ok := mem.objects.Delete(objectID.String())
	if ok {
		return 1, nil
	}
	return 0, nil
}

func (mem *Store) Create(opts *objects.CreateOpts) (objects.Writer, error) {
	return NewMemDataWriter(mem), nil
}

func (mem *Store) Used() int64 {
	return mem.used.Load()
}

func (mem *Store) Free() int64 {
	return mem.size - mem.used.Load()
}
