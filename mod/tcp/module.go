package tcp

import (
	"github.com/cryptopunkscc/astrald/net"
	"github.com/cryptopunkscc/astrald/node/infra"
	_net "net"
)

const ModuleName = "tcp"

type Module interface {
	infra.Dialer
	infra.Unpacker
	infra.Parser
	infra.EndpointLister
	ListenPort() int
}

type Endpoint interface {
	net.Endpoint
	IP() _net.IP
	Port() int
}
