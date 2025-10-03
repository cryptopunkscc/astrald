package utp

import (
	"github.com/cryptopunkscc/astrald/mod/exonet"
)

const ModuleName = "utp"

type Module interface {
	exonet.Dialer
	exonet.Unpacker
	exonet.Parser
	ListenPort() int
}
