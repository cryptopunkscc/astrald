package tor

import (
	"github.com/cryptopunkscc/astrald/node/infra"
)

const ModuleName = "tor"

type Module interface {
	infra.Dialer
	infra.Unpacker
	infra.Parser
	infra.EndpointLister
}
