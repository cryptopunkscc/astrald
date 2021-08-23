package share

import (
	"github.com/cryptopunkscc/astrald/components/shares/mem"
	"github.com/cryptopunkscc/astrald/node"
)

func init() {
	shared := mem.NewSharedFiles()
	sc := serviceContext{shared}
	_ = node.RegisterService(Port, sc.runLocal)
	_ = node.RegisterService(RemotePort, sc.runRemote)
}
