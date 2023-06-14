package agent

import (
	"context"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/net"
	"github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/node/link"
	"github.com/cryptopunkscc/astrald/query"
)

const serviceName = "sys.agent"

type Module struct {
	node node.Node
	log  *log.Logger
	ctx  context.Context
}

func (module *Module) RouteQuery(ctx context.Context, q query.Query, remoteWriter net.SecureWriteCloser) (net.SecureWriteCloser, error) {
	if q.Origin() != query.OriginLocal {
		return nil, link.ErrRejected
	}

	return query.Accept(q, remoteWriter, module.serve)
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
