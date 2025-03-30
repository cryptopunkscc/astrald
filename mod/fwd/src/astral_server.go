package fwd

import (
	"context"
	"github.com/cryptopunkscc/astrald/astral"
	"io"
	"strings"
)

type AstralServer struct {
	*Module
	serviceName string
	identity    *astral.Identity
	target      astral.Router
}

func NewAstralServer(mod *Module, serviceName string, target astral.Router) (*AstralServer, error) {
	var err error
	var identity = mod.node.Identity()
	var srv = &AstralServer{
		Module: mod,
		target: target,
	}

	if idx := strings.Index(serviceName, "@"); idx != -1 {
		name := serviceName[:idx]
		identity, err = mod.Dir.ResolveIdentity(name)
		if err != nil {
			return nil, err
		}

		serviceName = serviceName[idx+1:]
	}

	srv.identity = identity
	srv.serviceName = serviceName

	return srv, nil
}

func (srv *AstralServer) Run(ctx *astral.Context) error {
	var err = srv.AddRoute(srv.serviceName, srv)
	if err != nil {
		return err
	}
	defer srv.RemoveRoute(srv.serviceName)

	<-ctx.Done()

	return nil
}

func (srv *AstralServer) RouteQuery(ctx context.Context, q *astral.Query, w io.WriteCloser) (io.WriteCloser, error) {
	dst, err := srv.target.RouteQuery(ctx, q, w)
	if err != nil {
		return nil, err
	}

	return dst, nil
}

func (srv *AstralServer) Target() astral.Router {
	return srv.target
}

func (srv *AstralServer) String() string {
	return "astral://" + srv.serviceName
}
