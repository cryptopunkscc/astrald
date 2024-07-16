package tor

import (
	"github.com/cryptopunkscc/astrald/mod/exonet"
)

const ModuleName = "tor"

type Module interface {
	exonet.Dialer
	exonet.Unpacker
	exonet.Parser
}
