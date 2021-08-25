package init

import (
	"github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/services/push"
)

func init() {
	_ = node.RegisterService(push.Port, push.NewService().Run)
}
