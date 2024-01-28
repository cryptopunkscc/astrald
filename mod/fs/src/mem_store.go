package fs

import (
	"github.com/cryptopunkscc/astrald/data"
	"github.com/cryptopunkscc/astrald/mod/storage"
	"github.com/cryptopunkscc/astrald/node/events"
	"github.com/cryptopunkscc/astrald/sig"
	"sync/atomic"
)

var _ storage.Reader = &MemStore{}
var _ storage.Store = &MemStore{}

const DefaultMemStoreSize = 64 * 1024 * 1024 // 64MB

type MemStore struct {
	objects sig.Map[string, []byte]
	events  events.Queue
	used    atomic.Int64
	size    int64
}

func NewMemStore(events *events.Queue, size int64) *MemStore {
	var mem = &MemStore{size: DefaultMemStoreSize}
	if size != 0 {
		mem.size = size
	}
	mem.events.SetParent(events)
	return mem
}

func (mem *MemStore) Read(dataID data.ID, opts *storage.ReadOpts) (storage.DataReader, error) {
	bytes, found := mem.objects.Get(dataID.String())
	if !found {
		return nil, storage.ErrNotFound
	}

	return NewMemDataReader(bytes), nil
}

func (mem *MemStore) Store(opts *storage.StoreOpts) (storage.DataWriter, error) {
	return NewMemDataWriter(mem), nil
}

func (mem *MemStore) Used() int64 {
	return mem.used.Load()
}

func (mem *MemStore) Free() int64 {
	return mem.size - mem.used.Load()
}
