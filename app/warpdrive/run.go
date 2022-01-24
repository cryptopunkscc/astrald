package warpdrive

import (
	"github.com/cryptopunkscc/astrald/app/warpdrive/api"
	"github.com/cryptopunkscc/astrald/app/warpdrive/core"
	"github.com/cryptopunkscc/astrald/app/warpdrive/handle"
	"github.com/cryptopunkscc/astrald/app/warpdrive/service"
)

// Run warpdrive service with default core and handlers.
func (s Service) Run() {
	srv := service.Context(s)
	srv.Core = core.New(s.Config)
	srv.Setup()
	srv.Serve(handlers)
}

type Service service.Context

var handlers = service.Handlers{
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
