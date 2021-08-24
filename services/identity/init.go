package identity

import (
	"github.com/cryptopunkscc/astrald/node"
)

func init() {
	srv := service{}
	_ = node.RegisterService(Port, srv.runService)
}