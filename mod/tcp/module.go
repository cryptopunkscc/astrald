package tcp

import (
	"github.com/cryptopunkscc/astrald/net"
	"github.com/cryptopunkscc/astrald/node"
	_net "net"
)

const ModuleName = "tcp"

type Module interface {
	node.Dialer
	node.Unpacker
	node.Parser
	node.EndpointLister
	ListenPort() int
}

type Endpoint interface {
	net.Endpoint
	IP() _net.IP
	Port() int
}
