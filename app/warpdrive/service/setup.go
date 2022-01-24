package service

import (
	astral "github.com/cryptopunkscc/astrald/mod/apphost/client"
	"github.com/cryptopunkscc/astrald/mod/contacts"
	"github.com/cryptopunkscc/astrald/mod/id"
	"log"
)

func (srv *Context) Setup() *Context {
	srv.Logger = log.Default()
	srv.Core.Setup()
	srv.setupAstralApi()
	srv.setupIdentity()
	srv.setupPeers()
	return srv
}

func (srv *Context) setupAstralApi() {
	if srv.Api == nil {
		srv.Api = astral.Instance()
	}
}

func (srv *Context) setupIdentity() {
	identity, err := id.Query()
	if err != nil {
		srv.Panic("Cannot obtain node identity", err)
	}
	srv.Identity = identity.String()
}

func (srv *Context) setupPeers() {
	contactList, err := contacts.Query()
	if err != nil {
		srv.Println("Cannot obtain contacts", err)
		return
	}
	for _, contact := range contactList {
		srv.Peer().Update(contact.Id, "", "")
	}
}
