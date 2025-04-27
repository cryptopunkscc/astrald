package mem

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/object"
	"github.com/cryptopunkscc/astrald/sig"
	"sync/atomic"
)

var _ objects.Opener = &Store{}
var _ objects.Repository = &Store{}
var _ objects.Purger = &Store{}

const DefaultSize = 64 * 1024 * 1024 // 64MB

type Store struct {
	objects sig.Map[string, []byte]
	mod     objects.Module
	used    atomic.Int64
	size    int64
	name    string
}

func (mem *Store) Scan() (<-chan *object.ID, error) {
	ch := make(chan *object.ID)

	go func() {
		defer close(ch)

		for _, s := range mem.objects.Keys() {
			id, err := object.ParseID(s)
			if err != nil {
				continue
			}
			ch <- &id
		}
	}()

	return ch, nil
}

func NewMemStore(mod objects.Module, name string, size int64) *Store {
	var mem = &Store{mod: mod, size: DefaultSize, name: name}
	if size != 0 {
		mem.size = size
	}
	return mem
}

func (mem *Store) OpenObject(ctx *astral.Context, objectID object.ID, opts *objects.OpenOpts) (objects.Reader, error) {
	if !ctx.Zone().Is(astral.ZoneDevice) {
		return nil, astral.ErrZoneExcluded
	}

	bytes, found := mem.objects.Get(objectID.String())
	if !found {
		return nil, objects.ErrNotFound
	}

	return NewMemDataReader(bytes), nil
}

func (mem *Store) PurgeObject(objectID object.ID, opts *objects.PurgeOpts) (int, error) {
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

func (mem *Store) Name() string {
	return mem.name
}
