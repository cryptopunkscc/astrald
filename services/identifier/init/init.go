package init

import (
	"github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/services/identifier"
)

func init() {
	_ = node.RegisterService(identifier.Port, identifier.Run)
}
