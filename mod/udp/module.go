package udp

import (
	"github.com/cryptopunkscc/astrald/mod/exonet"
)

const ModuleName = "udp"

type Module interface {
	exonet.Dialer
	exonet.Unpacker
	exonet.Parser
	ListenPort() int
}
