package tcp

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/exonet"
)

const ModuleName = "tcp"

const (
	MethodNewEphemeralListener   = "tcp.new_ephemeral_listener"
	MethodCloseEphemeralListener = "tcp.close_ephemeral_listener"
)

type Module interface {
	exonet.Dialer
	exonet.Unpacker
	exonet.Parser
	ListenPort() int
	CreateEphemeralListener(ctx *astral.Context, port astral.Uint16, handler exonet.EphemeralHandler) error
}
