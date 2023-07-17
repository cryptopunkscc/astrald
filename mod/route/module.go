package route

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/mod/route/proto"
	"github.com/cryptopunkscc/astrald/net"
	"github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/node/assets"
	"github.com/cryptopunkscc/astrald/tasks"
	"io"
)

type Module struct {
	node   node.Node
	keys   assets.KeyStore
	log    *log.Logger
	config Config
}

func (m *Module) Run(ctx context.Context) error {
	return tasks.Group(
		&RouteService{Module: m},
	).Run(ctx)
}

func (m *Module) RouteVia(ctx context.Context, relay id.Identity, query net.Query, caller net.SecureWriteCloser) (target net.SecureWriteCloser, err error) {
	if query.Caller().PrivateKey() == nil {
		return nil, errors.New("caller private key missing")
	}

	routeConn, err := net.Route(ctx, m.node.Router(), net.NewQuery(m.node.Identity(), relay, RouteServiceName))
	if err != nil {
		return nil, err
	}

	var rpc = proto.New(routeConn)

	if !query.Caller().IsEqual(m.node.Identity()) {
		var cert = proto.NewRelayCert(query.Caller(), routeConn.LocalIdentity())

		if err = rpc.Shift(cert); err != nil {
			return nil, err
		}
	}

	err = rpc.Query(query.Target(), query.Query())
	if err != nil {
		return nil, err
	}

	if !routeConn.RemoteIdentity().IsEqual(query.Target()) {
		var cert proto.RelayCert
		if err := rpc.Decode(&cert); err != nil {
			return nil, err
		}

		if !cert.Identity.IsEqual(query.Target()) {
			return nil, errors.New("received invalid certificate")
		}

		if !cert.Relay.IsEqual(routeConn.RemoteIdentity()) {
			return nil, errors.New("received invalid certificate")
		}

		if err = cert.Verify(); err != nil {
			return nil, err
		}

		m.log.Logv(2, "target shifted to %v", cert.Identity)
	}

	if err = rpc.DecodeErr(); err != nil {
		return nil, err
	}

	go func() {
		io.Copy(caller, routeConn)
		caller.Close()
	}()

	return net.NewSecureWriteCloser(routeConn, query.Target()), nil
}
