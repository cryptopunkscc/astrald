package agent

import (
	"context"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/mod/dir"
	"github.com/cryptopunkscc/astrald/net"
	"github.com/cryptopunkscc/astrald/node"
)

const serviceName = "sys.agent"

type Module struct {
	node node.Node
	log  *log.Logger
	ctx  context.Context

	dir dir.Module
}

func (mod *Module) RouteQuery(ctx context.Context, query net.Query, caller net.SecureWriteCloser, hints net.Hints) (net.SecureWriteCloser, error) {
	if hints.Origin != net.OriginLocal {
		return nil, net.ErrRejected
	}

	return net.Accept(query, caller, mod.serve)
}

func (mod *Module) serve(conn net.SecureConn) {
	s := &Server{
		mod:  mod,
		conn: conn,
		log:  mod.log,
	}

	if err := s.Run(mod.ctx); err != nil {
		mod.log.Error("serve error: %s", err)
	}
}

func (mod *Module) Run(ctx context.Context) error {
	mod.ctx = ctx

	err := mod.node.LocalRouter().AddRoute(serviceName, mod)
	if err != nil {
		return err
	}
	defer mod.node.LocalRouter().RemoveRoute(serviceName)

	<-ctx.Done()

	return nil
}
