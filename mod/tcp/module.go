package tcp

import (
	"github.com/cryptopunkscc/astrald/node/infra"
)

const ModuleName = "tcp"

type Module interface {
	infra.Dialer
	infra.Unpacker
	infra.Parser
	infra.EndpointLister
	ListenPort() int
}
