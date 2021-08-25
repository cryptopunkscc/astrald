package init

import (
	"github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/services/sync"
)

func init() {
	_ = node.RegisterService(sync.Port, sync.Run)
}
