package warpdrive

import (
	"github.com/cryptopunkscc/astrald/app/warpdrive/api"
	"github.com/cryptopunkscc/astrald/app/warpdrive/handle"
	"github.com/cryptopunkscc/astrald/app/warpdrive/handler"
	"github.com/cryptopunkscc/astrald/app/warpdrive/service"
	astral "github.com/cryptopunkscc/astrald/mod/apphost/client"
)

type Client handle.Client

type Service handler.Context

var handlers = handler.Handlers{
	api.Send:        handle.ServiceSend,
	api.Accept:      handle.ServiceAccept,
	api.Reject:      handle.ServiceReject,
	api.SenPeers:    handle.SenderPeers,
	api.SenSend:     handle.SenderSend,
	api.SenStatus:   handle.SenderStatus,
	api.SenSent:     handle.SenderSent,
	api.SenEvents:   handle.SenderEvents,
	api.RecIncoming: handle.RecipientOffers,
	api.RecReceived: handle.RecipientReceived,
	api.RecAccept:   handle.RecipientAccept,
	api.RecReject:   handle.RecipientReject,
	api.RecUpdate:   handle.RecipientUpdate,
	api.RecEvents:   handle.RecipientEvents,
	api.CliQuery:    handle.CommandLine,
}

// Run warpdrive service with default core and handlers.
func (srv Service) Run() {
	srv.Core = service.Core(srv.Config)
	ctx := handler.Context(srv)
	ctx.Init()
	ctx.Serve(handlers)
}

// NewClient returns default warpdrive client.
func NewClient() Client {
	return Client(handle.NewClient(astral.Instance()))
}
