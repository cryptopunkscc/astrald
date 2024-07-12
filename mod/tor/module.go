package tor

import (
	"github.com/cryptopunkscc/astrald/node"
)

const ModuleName = "tor"

type Module interface {
	node.Dialer
	node.Unpacker
	node.Parser
	node.EndpointLister
}
