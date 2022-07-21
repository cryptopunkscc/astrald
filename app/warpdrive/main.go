package warpdrive

import (
	"github.com/cryptopunkscc/astrald/app/warpdrive/api"
	"github.com/cryptopunkscc/astrald/app/warpdrive/handle"
	"github.com/cryptopunkscc/astrald/app/warpdrive/handler"
	"github.com/cryptopunkscc/astrald/app/warpdrive/setup"
	"github.com/cryptopunkscc/astrald/lib/astral"
)

type Client handle.Client

type Service handler.Context

var handlers = handler.Handlers{{
	api.QueryPeers:     handle.Peers,
	api.QuerySend:      handle.Send,
	api.QueryAccept:    handle.Download,
	api.QueryUpdate:    handle.Update,
	api.QuerySubscribe: handle.Subscribe,
	api.QueryStatus:    handle.Status,
	api.QueryOffers:    handle.Offers,
	api.QueryOffer:     handle.Receive,
	api.QueryFiles:     handle.Upload,
	api.QueryCli:       handle.Cli,
}, {
	api.Port: handle.Ping,
}}

// Run warpdrive service with default core and handlers.
func (srv Service) Run() {
	ctx := handler.Context(srv)
	setup.Core(&ctx.Core)
	setup.Context(&ctx)
	ctx.Serve(handlers)
}

// NewClient returns default warpdrive client.
func NewClient() Client {
	return Client(handle.NewClient(astral.AppHostAdapter()))
}
