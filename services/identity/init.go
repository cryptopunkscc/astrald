package identity

import (
	"github.com/cryptopunkscc/astrald/components/uid/file"
	"github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/services"
)

func init() {
	c := Context{ids: file.NewIdentities(services.AstralHome)}
	_ = node.RegisterService(Port, c.runService)
}