package user

import (
	"errors"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/data"
	"github.com/cryptopunkscc/astrald/mod/relay"
	"github.com/cryptopunkscc/astrald/mod/storage"
	"github.com/cryptopunkscc/astrald/node/router"
)

func (mod *Module) AddIdentity(identity id.Identity) error {
	var err = mod.addIdentity(identity)
	if err != nil {
		return err
	}

	var tx = mod.db.Create(&dbIdentity{Identity: identity.PublicKeyHex()})

	return tx.Error
}

func (mod *Module) addIdentity(identity id.Identity) error {
	if _, found := mod.identities.Get(identity.PublicKeyHex()); found {
		return errors.New("identity already added")
	}

	var err error
	var i = &Identity{
		identity: identity,
		cert:     nil,
		routes:   router.NewPrefixRouter(false),
	}
	i.routes.EnableParams = true

	// get user certificate for local relay
	i.cert, err = mod.relay.ReadCert(&relay.FindOpts{
		TargetID:  identity,
		RelayID:   mod.node.Identity(),
		Direction: relay.Both,
	})

	switch {
	case errors.Is(err, relay.ErrCertNotFound):
		mod.log.Logv(1, "relay certificate for %v not found, generating...", identity)

		certID, err := mod.relay.MakeCert(identity, mod.node.Identity(), relay.Both, 0)
		if err != nil {
			mod.log.Error("error generating relay certificate for %v: %v", identity, err)
		}

		i.cert, err = mod.storage.ReadAll(certID, &storage.ReadOpts{Virtual: true})
		if err != nil {
			return err
		}

	case err != nil:
		return err
	}

	if !mod.identities.Set(i.identity.PublicKeyHex(), i) {
		return errors.New("identity already added")
	}

	mod.log.Infov(1, "added user %v@%v (cert %v)",
		i.identity,
		mod.node.Identity(),
		data.Resolve(i.cert),
	)

	i.routes.AddRoute(userProfileServiceName, mod.profileService)
	i.routes.AddRoute(notifyServiceName, mod.notifyService)

	mod.node.Router().AddRoute(id.Anyone, i.identity, mod, 100)

	if mod.admin != nil {
		mod.admin.AddAdmin(i.identity)
	}

	return nil
}
