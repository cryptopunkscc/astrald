package main

import (
	"github.com/cryptopunkscc/astrald/data"
	astral "github.com/cryptopunkscc/astrald/mod/apphost/client"
	"github.com/cryptopunkscc/astrald/proto/block"
	"github.com/cryptopunkscc/astrald/proto/store"
	"log"
)

type AuthStore struct {
	conn       *astral.Conn
	store      store.Store
	remoteName string
}

func NewAuthStore(conn *astral.Conn, store store.Store) *AuthStore {
	s := &AuthStore{
		conn:       conn,
		store:      store,
		remoteName: conn.RemoteIdentity().String(),
	}

	if info, err := astral.NodeInfo(conn.RemoteIdentity()); err == nil {
		s.remoteName = info.Name
	}

	return s
}

func (store *AuthStore) Open(id data.ID, flags uint32) (block.Block, error) {
	log.Printf("%s open %s (flags %04x)\n", store.remoteName, id.String(), flags)

	return store.store.Open(id, flags)
}

func (store *AuthStore) Create(alloc uint64) (block.Block, string, error) {
	return store.store.Create(alloc)
}
