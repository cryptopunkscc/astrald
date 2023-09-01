package agent

import (
	"context"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/net"
	"github.com/cryptopunkscc/astrald/node"
)

const serviceName = "sys.agent"

type Module struct {
	node node.Node
	log  *log.Logger
	ctx  context.Context
}

func (module *Module) RouteQuery(ctx context.Context, query net.Query, caller net.SecureWriteCloser, hints net.Hints) (net.SecureWriteCloser, error) {
	if hints.Origin != net.OriginLocal {
		return nil, net.ErrRejected
	}

	return net.Accept(query, caller, module.serve)
}

func (module *Module) serve(conn net.SecureConn) {
	s := &Server{
		node: module.node,
		conn: conn,
		log:  module.log,
	}

	if err := s.Run(module.ctx); err != nil {
		module.log.Error("serve error: %s", err)
	}
}

func (module *Module) Run(ctx context.Context) error {
	module.ctx = ctx

	service, err := module.node.Services().Register(ctx, module.node.Identity(), serviceName, module)
	if err != nil {
		return err
	}

	<-service.Done()

	return nil
}
