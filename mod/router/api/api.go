package router

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/net"
	"github.com/cryptopunkscc/astrald/node/modules"
)

const ModuleName = "router"

type API interface {
	SetRouter(target id.Identity, router id.Identity)
	GetRouter(target id.Identity) id.Identity
	Reroute(nonce net.Nonce, router net.Router) error
}

func Load(node modules.Node) (API, error) {
	api, ok := node.Modules().Find(ModuleName).(API)
	if !ok {
		return nil, modules.ErrNotFound
	}
	return api, nil
}
