package identifier

import "github.com/cryptopunkscc/astrald/node"

func init() {
	_ = node.RegisterService(Port, run)
}
