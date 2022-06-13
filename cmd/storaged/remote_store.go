package main

import (
	"github.com/cryptopunkscc/astrald/data"
	"github.com/cryptopunkscc/astrald/lib/astral"
	_block "github.com/cryptopunkscc/astrald/proto/block"
	"github.com/cryptopunkscc/astrald/proto/store"
)

var _ store.Store = &RemoteStore{}

type RemoteStore struct {
	sources []string
}

func NewRemoteStore(sources []string) *RemoteStore {
	return &RemoteStore{sources: sources}
}

func (r RemoteStore) Open(id data.ID, flags uint32) (_block.Block, error) {
	for _, source := range r.sources {
		conn, err := astral.DialName(source, "storage")
		if err != nil {
			continue
		}

		remoteStore := store.Bind(conn)
		block, err := remoteStore.Open(id, flags&^uint32(store.OpenRemote))
		if err == nil {
			return block, nil
		}
		conn.Close()
	}

	return nil, store.ErrNotFound
}

func (r RemoteStore) Create(alloc uint64) (_block.Block, string, error) {
	return nil, "", store.ErrUnsupported
}
