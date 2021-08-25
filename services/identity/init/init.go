package init

import (
	"github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/services/identity"
)

func init() {
	srv := identity.Service{}
	_ = node.RegisterService(identity.Port, srv.Run)
}