package init

import (
	"github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/services/warpdrive"
)

func init() {
	_ = node.RegisterService(warpdrive.PortLocal, warpdrive.RunLocal)
	_ = node.RegisterService(warpdrive.PortRemote, warpdrive.RunRemote)
}
