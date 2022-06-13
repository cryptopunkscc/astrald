package main

import (
	"github.com/cryptopunkscc/astrald/data"
	astral2 "github.com/cryptopunkscc/astrald/lib/astral"
	"github.com/cryptopunkscc/astrald/proto/block"
	"github.com/cryptopunkscc/astrald/proto/store"
	"log"
)

type AuthStore struct {
	conn       *astral2.Conn
	store      store.Store
	remoteName string
}

func NewAuthStore(conn *astral2.Conn, store store.Store) *AuthStore {
	s := &AuthStore{
		conn:       conn,
		store:      store,
		remoteName: conn.RemoteIdentity().String(),
	}

	if info, err := astral2.NodeInfo(conn.RemoteIdentity()); err == nil {
		s.remoteName = info.Name
	}

	return s
}

func (store *AuthStore) Open(id data.ID, flags uint32) (block.Block, error) {
	log.Printf("%s open %s (flags 0x%04x)\n", store.remoteName, id.String(), flags)

	return store.store.Open(id, flags)
}

func (store *AuthStore) Create(alloc uint64) (block.Block, string, error) {
	log.Printf("%s create (alloc %d)\n", store.remoteName, alloc)

	create, s, err := store.store.Create(alloc)
	if err == nil {
		log.Printf("%s created %s\n", store.remoteName, s)
	} else {
		log.Printf("%s error %s\n", store.remoteName, err.Error())
	}
	return create, s, err
}
