package init

import (
	"github.com/cryptopunkscc/astrald/components/shares/mem"
	"github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/services/share"
)

func init() {
	shared := mem.NewSharedFiles()
	srv := share.NewService(shared)
	_ = node.RegisterService(share.Port, srv.RunLocal)
	_ = node.RegisterService(share.RemotePort, srv.RunRemote)
}
