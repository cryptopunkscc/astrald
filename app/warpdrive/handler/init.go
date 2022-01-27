package handler

import (
	"github.com/cryptopunkscc/astrald/app/warpdrive/service"
	astral "github.com/cryptopunkscc/astrald/mod/apphost/client"
	"github.com/cryptopunkscc/astrald/mod/contacts"
	"github.com/cryptopunkscc/astrald/mod/id"
)

func (ctx *Context) Init() *Context {
	ctx.initAstralApi()
	ctx.initIdentity()
	ctx.initPeers()
	return ctx
}

func (ctx *Context) initAstralApi() {
	if ctx.Api == nil {
		ctx.Api = astral.Instance()
	}
}

func (ctx *Context) initIdentity() {
	identity, err := id.Query()
	if err != nil {
		ctx.Panic("Cannot obtain node identity", err)
	}
	ctx.Identity = identity.String()
}

func (ctx *Context) initPeers() {
	contactList, err := contacts.Query()
	if err != nil {
		ctx.Println("Cannot obtain contacts", err)
		return
	}
	peerService := service.Peer(ctx.Core)
	for _, contact := range contactList {
		peerService.Update(contact.Id, "", "")
	}
}
