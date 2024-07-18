package shares

import (
	"context"
	"encoding/json"
	"github.com/cryptopunkscc/astrald/lib/routers"
	"github.com/cryptopunkscc/astrald/mod/sets/sync"
	"github.com/cryptopunkscc/astrald/astral"
)

const readServiceName = "shares.read"
const notifyServiceName = "shares.notify"

type JSONDescriptor struct {
	Type string
	Data json.RawMessage
}

type Provider struct {
	*Module
	router *routers.PrefixRouter
}

func (srv *Provider) Run(ctx context.Context) error {
	return nil
}

func NewProvider(mod *Module) *Provider {
	var srv = &Provider{
		Module: mod,
		router: routers.NewPrefixRouter(true),
	}

	srv.router.EnableParams = true

	srv.router.AddRouteFunc("shares.sync", srv.Sync)
	srv.router.AddRouteFunc("shares.notify", srv.Notify)

	return srv
}

func (srv *Provider) RouteQuery(ctx context.Context, query astral.Query, caller astral.SecureWriteCloser, hints astral.Hints) (astral.SecureWriteCloser, error) {
	return srv.router.RouteQuery(ctx, query, caller, hints)
}

func (srv *Provider) Sync(ctx context.Context, query astral.Query, caller astral.SecureWriteCloser, hints astral.Hints) (astral.SecureWriteCloser, error) {
	set, err := srv.openExportSet(caller.Identity())
	if err != nil {
		return astral.Reject()
	}

	p := sync.NewProvider(set)

	return p.RouteQuery(ctx, query, caller, hints)
}

func (srv *Provider) Notify(ctx context.Context, query astral.Query, caller astral.SecureWriteCloser, hints astral.Hints) (astral.SecureWriteCloser, error) {
	remoteShare, err := srv.FindRemoteShare(query.Target(), query.Caller())
	if err != nil {
		return astral.Reject()
	}

	return astral.Accept(query, caller, func(conn astral.Conn) {
		conn.Close()
		srv.tasks <- func(ctx context.Context) {
			remoteShare.Sync(ctx)
		}
	})
}
