package identity

import (
	"github.com/cryptopunkscc/astrald/node"
)

func init() {
	c := Context{}
	_ = node.RegisterService(Port, c.runService)
}