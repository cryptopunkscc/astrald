package share

import (
	"github.com/cryptopunkscc/astrald/components/shares/mem"
	"github.com/cryptopunkscc/astrald/node"
)

func init() {
	shared := mem.NewSharedFiles()
	service := service{shared}
	_ = node.RegisterService(Port, service.runLocal)
	_ = node.RegisterService(RemotePort, service.runRemote)
}
