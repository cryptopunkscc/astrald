package tcp

import (
	"github.com/cryptopunkscc/astrald/mod/exonet"
	_net "net"
)

const ModuleName = "tcp"

type Module interface {
	exonet.Dialer
	exonet.Unpacker
	exonet.Parser
	ListenPort() int
}

type Endpoint interface {
	exonet.Endpoint
	IP() _net.IP
	Port() int
}
