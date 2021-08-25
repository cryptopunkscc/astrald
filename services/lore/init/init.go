package init

import (
	"github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/services/lore"
)

func init() {
	_ = node.RegisterService(lore.Port, lore.Run)
}
