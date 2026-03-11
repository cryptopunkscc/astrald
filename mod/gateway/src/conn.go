package gateway

import (
	"github.com/cryptopunkscc/astrald/mod/exonet"
	"github.com/cryptopunkscc/astrald/mod/gateway"
)

type gwConn struct {
	exonet.Conn
	remote *gateway.Endpoint
}

func (c *gwConn) RemoteEndpoint() exonet.Endpoint {
	return c.remote
}
