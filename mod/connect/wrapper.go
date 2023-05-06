package connect

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/net"
	"github.com/cryptopunkscc/astrald/node/infra/drivers/gw"
	"io"
)

var _ net.Conn = &wrapper{}

type wrapper struct {
	io.ReadWriteCloser
	local    id.Identity
	remote   id.Identity
	outbound bool
}

func (w wrapper) Outbound() bool {
	return w.outbound
}

func (w wrapper) LocalEndpoint() net.Endpoint {
	return gw.NewEndpoint(w.remote, w.local)
}

func (w wrapper) RemoteEndpoint() net.Endpoint {
	return gw.NewEndpoint(w.remote, w.local)
}
