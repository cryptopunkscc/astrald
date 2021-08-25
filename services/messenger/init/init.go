package init

import (
	"github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/services/messenger"
)

func init() {
	_ = node.RegisterService(messenger.Port, messenger.Run)
}