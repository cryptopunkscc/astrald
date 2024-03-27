package user

import (
	"errors"
	"fmt"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/mod/relay"
	"github.com/cryptopunkscc/astrald/mod/storage"
	"github.com/cryptopunkscc/astrald/mod/user"
	"gorm.io/gorm"
)

type LocalUser struct {
	mod      *Module
	identity id.Identity
	cert     []byte
}

func NewLocalUser(mod *Module, identity id.Identity) *LocalUser {
	i := &LocalUser{
		mod:      mod,
		identity: identity,
	}
	return i
}

func (mod *Module) LocalUser() user.LocalUser {
	if mod.localUser == nil {
		return nil
	}
	return mod.localUser
}

func (mod *Module) SetLocalUser(identity id.Identity) error {
	var err = mod.db.Session(&gorm.Session{AllowGlobalUpdate: true}).Delete(&dbIdentity{}).Error
	if err != nil {
		return err
	}

	mod.localUser = nil

	if identity.IsZero() {
		return nil
	}

	err = mod.db.Create(&dbIdentity{Identity: identity}).Error
	if err != nil {
		return err
	}

	localUser := NewLocalUser(mod, identity)
	if err = localUser.load(); err != nil {
		return err
	}

	mod.localUser = localUser

	mod.log.Info("local user set to %v", identity)

	return nil
}

func (i *LocalUser) load() error {
	var err error

	err = i.loadCert()
	if err != nil {
		err = i.createCert()
		if err != nil {
			return fmt.Errorf("create cert: %w", err)
		}
	}

	return nil
}

func (i *LocalUser) loadCert() error {
	var err error

	// get user certificate for local relay
	i.cert, err = i.mod.relay.ReadCert(&relay.FindOpts{
		TargetID:  i.identity,
		RelayID:   i.mod.node.Identity(),
		Direction: relay.Both,
	})

	if errors.Is(err, relay.ErrCertNotFound) {
		i.mod.log.Logv(1, "relay certificate for %v not found, generating...", i.identity)
		err = i.createCert()
		if err != nil {
			i.mod.log.Error("error generating relay certificate for %v: %v", i.identity, err)
		}
	}

	return err
}

func (i *LocalUser) createCert() error {
	certID, err := i.mod.relay.MakeCert(i.identity, i.mod.node.Identity(), relay.Both, 0)
	if err != nil {
		return err
	}

	i.cert, err = i.mod.storage.ReadAll(certID, &storage.OpenOpts{Virtual: true})
	if err != nil {
		return err
	}

	return nil
}

func (i *LocalUser) Identity() id.Identity {
	return i.identity
}

func (i *LocalUser) Cert() []byte {
	return i.cert
}
