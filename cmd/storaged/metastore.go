package main

import (
	"github.com/cryptopunkscc/astrald/data"
	_block "github.com/cryptopunkscc/astrald/proto/block"
	"github.com/cryptopunkscc/astrald/proto/store"
)

var _ store.Store = &MetaStore{}

type MetaStore struct {
	local  store.Store
	remote store.Store
}

func (s MetaStore) Open(id data.ID, flags uint32) (_block.Block, error) {
	block, err := s.local.Open(id, flags)
	if err == nil {
		return block, err
	}

	if (s.remote != nil) && (flags&store.OpenRemote != 0) {
		block, err := s.remote.Open(id, flags)
		if err == nil {
			return block, err
		}
	}

	return block, err
}

func (s MetaStore) Create(alloc uint64) (_block.Block, string, error) {
	return s.local.Create(alloc)
}

func NewMetaStore(dataDir string, sources []string) *MetaStore {
	return &MetaStore{
		local:  NewDirStore(dataDir),
		remote: NewRemoteStore(sources),
	}
}
