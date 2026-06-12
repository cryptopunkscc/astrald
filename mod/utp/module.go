package utp

import (
	"github.com/cryptopunkscc/astrald/mod/exonet"
)

const ModuleName = "utp"

// Module is the public contract for the UTP transport.
type Module interface {
	exonet.Dialer
	exonet.Unpacker
	exonet.Parser
	ListenPort() int
}
