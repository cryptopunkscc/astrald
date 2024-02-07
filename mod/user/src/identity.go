package user

import (
	"errors"
	"fmt"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/mod/relay"
	"github.com/cryptopunkscc/astrald/mod/sets"
	"github.com/cryptopunkscc/astrald/mod/storage"
	"github.com/cryptopunkscc/astrald/node/router"
)

type Identity struct {
	mod         *Module
	identity    id.Identity
	row         *dbIdentity
	cert        []byte
	routes      *router.PrefixRouter
	rootSet     sets.Union
	explicitSet sets.Set
}

func NewIdentity(mod *Module, identity id.Identity) *Identity {
	i := &Identity{
		mod:      mod,
		identity: identity,
		routes:   router.NewPrefixRouter(false),
	}
	i.routes.EnableParams = true
	return i
}

func (i *Identity) init() error {
	var err error

	err = i.loadDb()
	if err != nil {
		err = i.createDb()
		if err != nil {
			return fmt.Errorf("init: db: %w", err)
		}
	}

	err = i.loadCert()
	if err != nil {
		err = i.createCert()
		if err != nil {
			return fmt.Errorf("init: cert: %w", err)
		}
	}

	err = i.loadSets()
	if err != nil {
		err = i.createSets()
		if err != nil {
			return fmt.Errorf("init: sets: %w", err)
		}
	}

	return nil
}

func (i *Identity) destroy() error {
	i.destroySets()
	i.destroyDb()
	return nil
}

func (i *Identity) loadDb() error {
	var err error

	err = i.mod.db.
		Where("identity = ?", i.identity).
		First(&i.row).Error

	return err
}

func (i *Identity) createDb() error {
	i.row = &dbIdentity{
		Identity: i.identity,
	}

	return i.mod.db.Create(&i.row).Error
}

func (i *Identity) destroyDb() error {
	return i.mod.db.
		Model(&dbIdentity{}).
		Delete("identity = ?", i.identity).Error
}

func (i *Identity) loadCert() error {
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

func (i *Identity) createCert() error {
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

func (i *Identity) loadSets() error {
	var err error

	if i.row.SetName == "" {
		return errors.New("set name is empty")
	}

	i.rootSet, err = i.mod.sets.OpenUnion(i.row.SetName, false)
	if err != nil {
		return err
	}

	var mainSetName = "." + i.row.SetName + ".main"
	i.explicitSet, err = i.mod.sets.Open(mainSetName, false)
	if err != nil {
		return err
	}

	return nil
}

func (i *Identity) createSets() error {
	var err error

	// create root set
	var rootSetName = "user." + i.identity.PublicKeyHex()
	i.rootSet, err = i.mod.sets.CreateUnion(rootSetName)
	if err != nil {
		return err
	}
	i.rootSet.SetDisplayName(
		fmt.Sprintf("{{%s}}'s data", i.identity.PublicKeyHex()),
	)

	// update database row
	i.row.SetName = rootSetName
	err = i.mod.db.
		Model(i.row).
		Update("set_name", rootSetName).Error
	if err != nil {
		i.mod.log.Error("createSets(): update db: %v", err)
	}

	// create main set
	var mainSetName = "." + rootSetName + ".main"
	i.explicitSet, err = i.mod.sets.Create(mainSetName)
	if err != nil {
		i.mod.log.Error("createSets(): create main db: %v", err)
	}
	err = i.rootSet.AddSubset(mainSetName)
	if err != nil {
		i.mod.log.Error("createSets(): add subset: %v", err)
	}

	share, err := i.mod.shares.LocalShare(i.identity, true)
	if err != nil {
		i.mod.log.Error("createSets(): local share: %v", err)
		return nil
	}

	err = share.AddSet(rootSetName)
	if err != nil {
		i.mod.log.Error("createSets(): add set: %v", err)
	}

	return nil
}

func (i *Identity) destroySets() error {
	if i.row.SetName == "" {
		return errors.New("set name is empty")
	}

	var mainSetName = "." + i.row.SetName + ".main"
	set, _ := i.mod.sets.Open(mainSetName, false)
	set.Delete()

	set, _ = i.mod.sets.Open(i.row.SetName, false)
	set.Delete()

	return nil
}

func (i *Identity) Identity() id.Identity {
	return i.identity
}

func (i *Identity) Cert() []byte {
	return i.cert
}

func (i *Identity) Routes() *router.PrefixRouter {
	return i.routes
}
