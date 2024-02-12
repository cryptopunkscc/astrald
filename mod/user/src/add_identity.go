package user

import (
	"fmt"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/data"
)

func (mod *Module) AddIdentity(identity id.Identity) error {
	return mod.addIdentity(identity)
}

func (mod *Module) addIdentity(identity id.Identity) error {
	if _, found := mod.identities.Get(identity.PublicKeyHex()); found {
		return errIdentityAlreadyAdded
	}

	var err error
	var i = NewIdentity(mod, identity)

	err = i.init()
	if err != nil {
		return fmt.Errorf("user init: %w", err)
	}

	if _, ok := mod.identities.Set(i.identity.PublicKeyHex(), i); !ok {
		return errIdentityAlreadyAdded
	}

	mod.log.Infov(1, "added user %v@%v (cert %v)",
		i.identity,
		mod.node.Identity(),
		data.Resolve(i.cert),
	)

	i.routes.AddRoute(userProfileServiceName, mod.profileService)
	i.routes.AddRoute(notifyServiceName, mod.notifyService)

	mod.node.Router().AddRoute(id.Anyone, i.identity, mod, 100)

	return nil
}
