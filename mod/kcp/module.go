package kcp

import (
	"github.com/cryptopunkscc/astrald/mod/exonet"
)

const ModuleName = "kcp"

type Module interface {
	exonet.Dialer
	exonet.Unpacker
	exonet.Parser
	ListenPort() int
}
