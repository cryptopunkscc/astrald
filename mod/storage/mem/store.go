package mem

import (
	"github.com/cryptopunkscc/astrald/data"
	"github.com/cryptopunkscc/astrald/mod/storage"
	"github.com/cryptopunkscc/astrald/node/events"
	"github.com/cryptopunkscc/astrald/sig"
	"sync/atomic"
)

var _ storage.Opener = &Store{}
var _ storage.Creator = &Store{}
var _ storage.Purger = &Store{}

const DefaultSize = 64 * 1024 * 1024 // 64MB

type Store struct {
	objects sig.Map[string, []byte]
	events  events.Queue
	used    atomic.Int64
	size    int64
}

func NewMemStore(events *events.Queue, size int64) *Store {
	var mem = &Store{size: DefaultSize}
	if size != 0 {
		mem.size = size
	}
	mem.events.SetParent(events)
	return mem
}

func (mem *Store) Open(dataID data.ID, opts *storage.OpenOpts) (storage.Reader, error) {
	bytes, found := mem.objects.Get(dataID.String())
	if !found {
		return nil, storage.ErrNotFound
	}

	return NewMemDataReader(bytes), nil
}

func (mem *Store) Purge(dataID data.ID, opts *storage.PurgeOpts) (int, error) {
	_, ok := mem.objects.Delete(dataID.String())
	if ok {
		return 1, nil
	}
	return 0, nil
}

func (mem *Store) Create(opts *storage.CreateOpts) (storage.Writer, error) {
	return NewMemDataWriter(mem), nil
}

func (mem *Store) Used() int64 {
	return mem.used.Load()
}

func (mem *Store) Free() int64 {
	return mem.size - mem.used.Load()
}
