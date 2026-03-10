package tcp

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/exonet"
)

const ModuleName = "tcp"

type Module interface {
	exonet.Dialer
	exonet.Unpacker
	exonet.Parser
	ListenPort() int

	NewServer(port astral.Uint16, onAccept exonet.EphemeralHandler) exonet.EphemeralListener
}
