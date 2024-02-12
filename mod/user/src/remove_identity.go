package user

import (
	"errors"
	"github.com/cryptopunkscc/astrald/auth/id"
)

func (mod *Module) RemoveIdentity(identity id.Identity) error {
	i, ok := mod.identities.Delete(identity.PublicKeyHex())
	if !ok {
		return errors.New("identity not found")
	}

	mod.node.Router().RemoveRoute(id.Anyone, identity, mod)

	return i.destroy()
}
