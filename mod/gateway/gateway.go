package gateway

import (
	"context"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/cslq"
	"github.com/cryptopunkscc/astrald/debug"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/net"
	"github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/node/infra/drivers/gw"
	"github.com/cryptopunkscc/astrald/streams"
)

const queryConnect = "connect"

type Gateway struct {
	node node.Node
	log  *log.Logger
	ctx  context.Context
}

func (module *Gateway) Run(ctx context.Context) error {
	module.ctx = ctx

	service, err := module.node.Services().Register(ctx, module.node.Identity(), gw.ServiceName, module)
	if err != nil {
		return err
	}

	<-service.Done()

	return nil
}

func (module *Gateway) RouteQuery(ctx context.Context, query net.Query, caller net.SecureWriteCloser, hints net.Hints) (net.SecureWriteCloser, error) {
	return net.Accept(query, caller, func(conn net.SecureConn) {
		defer debug.SaveLog(func(p any) {
			module.log.Error("gateway panicked: %v", p)
		})

		module.handleQuery(conn)
	})
}

func (module *Gateway) handleQuery(conn net.SecureConn) error {
	var err error
	var c = cslq.NewEndec(conn)
	var cookie string

	err = c.Decodef("[c]c", &cookie)
	if err != nil {
		return err
	}

	nodeID, err := id.ParsePublicKeyHex(cookie)
	if err != nil {
		module.log.Errorv(2, "invalid request from %v: malformed target identity", conn.RemoteIdentity())
		c.Encodef("c", false)
		conn.Close()
		return err
	}

	out, err := net.Route(module.ctx, module.node.Network(), net.NewQuery(module.node.Identity(), nodeID, queryConnect))
	if err != nil {
		conn.Close()
		return err
	}

	c.Encodef("c", true)

	l, r, err := streams.Join(conn, out)

	module.log.Logv(1, "conn for %s done (bytes read %d written %d)", nodeID, l, r)

	return err
}
