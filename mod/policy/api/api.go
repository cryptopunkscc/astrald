package policy

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/node/modules"
)

const ModuleName = "policy"

type API interface {
	AddAlwaysLinkedIdentity(identity id.Identity) error
	RemoveAlwaysLinkedIdentity(identity id.Identity) error
}

func Load(node modules.Node) (API, error) {
	api, ok := node.Modules().Find(ModuleName).(API)
	if !ok {
		return nil, modules.ErrNotFound
	}
	return api, nil
}
