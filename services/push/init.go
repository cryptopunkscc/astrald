package push

import (
	"github.com/cryptopunkscc/astrald/node"
)

func init() {
	_ = node.RegisterService(Port, NewService().Run)
}
