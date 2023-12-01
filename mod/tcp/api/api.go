package tcp

import (
	"github.com/cryptopunkscc/astrald/node/infra"
)

type API interface {
	infra.Dialer
	infra.Unpacker
	infra.Parser
	infra.EndpointLister
	ListenPort() int
}
