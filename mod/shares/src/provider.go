package shares

import (
	"context"
	"encoding/json"
	"github.com/cryptopunkscc/astrald/core"
	"github.com/cryptopunkscc/astrald/mod/sets/sync"
	"github.com/cryptopunkscc/astrald/net"
)

const readServiceName = "shares.read"
const notifyServiceName = "shares.notify"

type JSONDescriptor struct {
	Type string
	Data json.RawMessage
}

type Provider struct {
	*Module
	router *core.PrefixRouter
}

func (srv *Provider) Run(ctx context.Context) error {
	return nil
}

func NewProvider(mod *Module) *Provider {
	var srv = &Provider{
		Module: mod,
		router: core.NewPrefixRouter(true),
	}

	srv.router.EnableParams = true

	srv.router.AddRouteFunc("shares.sync", srv.Sync)
	srv.router.AddRouteFunc("shares.notify", srv.Notify)

	return srv
}

func (srv *Provider) RouteQuery(ctx context.Context, query net.Query, caller net.SecureWriteCloser, hints net.Hints) (net.SecureWriteCloser, error) {
	return srv.router.RouteQuery(ctx, query, caller, hints)
}

func (srv *Provider) Sync(ctx context.Context, query net.Query, caller net.SecureWriteCloser, hints net.Hints) (net.SecureWriteCloser, error) {
	set, err := srv.openExportSet(caller.Identity())
	if err != nil {
		return net.Reject()
	}

	p := sync.NewProvider(set)

	return p.RouteQuery(ctx, query, caller, hints)
}

func (srv *Provider) Notify(ctx context.Context, query net.Query, caller net.SecureWriteCloser, hints net.Hints) (net.SecureWriteCloser, error) {
	remoteShare, err := srv.FindRemoteShare(query.Target(), query.Caller())
	if err != nil {
		return net.Reject()
	}

	return net.Accept(query, caller, func(conn net.SecureConn) {
		conn.Close()
		srv.tasks <- func(ctx context.Context) {
			remoteShare.Sync(ctx)
		}
	})
}
