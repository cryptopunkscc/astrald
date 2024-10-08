package user

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/core/assets"
	"github.com/cryptopunkscc/astrald/lib/routers"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/mod/user"
	"github.com/cryptopunkscc/astrald/object"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

const assetLocalContract = "mod.user.local_contract"
const defaultContractValidity = 24 * time.Hour

var _ user.Module = &Module{}

type Module struct {
	Deps
	*routers.PathRouter
	config Config
	node   astral.Node
	log    *log.Logger
	assets assets.Assets
	db     *gorm.DB
	userID *astral.Identity
}

func (mod *Module) Run(ctx context.Context) error {
	<-ctx.Done()
	return nil
}

func (mod *Module) Nodes(userID *astral.Identity) (nodes []*astral.Identity) {
	err := mod.db.
		Model(&dbNodeContract{}).
		Where("expires_at > ?", time.Now()).
		Where("user_id = ?", userID).
		Select("node_id").
		Find(&nodes).
		Error

	if err != nil {
		mod.log.Errorv(1, "db error: %v", err)
	}

	return
}

func (mod *Module) Owner(nodeID *astral.Identity) (userID *astral.Identity) {
	err := mod.db.
		Model(&dbNodeContract{}).
		Where("expires_at > ?", time.Now().UTC()).
		Where("node_id = ?", nodeID).
		Select("user_id").
		Find(&userID).
		Error

	if err != nil {
		mod.log.Errorv(1, "db error: %v", err)
	}

	return
}

func (mod *Module) UserID() *astral.Identity {
	return mod.userID
}

func (mod *Module) SetUserID(userID *astral.Identity) error {
	err := mod.setUserID(userID)
	if err != nil {
		return err
	}

	return mod.storeUserID(userID)
}

func (mod *Module) setUserID(userID *astral.Identity) error {
	mod.userID = userID

	mod.log.Info("user identity set to %v", mod.userID)

	return nil
}

func (mod *Module) AddContact(userID *astral.Identity) error {
	return mod.db.Clauses(clause.OnConflict{DoNothing: true}).Create(&dbContact{
		UserID: userID,
	}).Error
}

func (mod *Module) RemoveContact(userID *astral.Identity) error {
	return mod.db.Delete(&dbContact{UserID: userID}).Error
}

func (mod *Module) IsContact(userID *astral.Identity) (b bool) {
	if userID.IsZero() {
		return
	}
	mod.db.
		Model(&dbContact{}).
		Where("user_id = ?", userID).
		Select("count(*) > 0").
		First(&b)
	return
}

func (mod *Module) Contacts() (contacts []*astral.Identity) {
	mod.db.
		Model(&dbContact{}).
		Select("user_id").
		Find(&contacts)
	return
}

func (mod *Module) ContractExists(contractID object.ID) (b bool) {
	mod.db.
		Model(&dbNodeContract{}).
		Where("object_id = ?", contractID).
		Select("count(*) > 0").
		First(&b)
	return
}

func (mod *Module) findContractID(userID *astral.Identity, nodeID *astral.Identity) (contractID object.ID, err error) {
	err = mod.db.
		Model(&dbNodeContract{}).
		Where("user_id = ? AND node_id = ? AND expires_at > ?", userID, nodeID, time.Now().UTC()).
		Order("expires_at DESC").
		Select("object_id").
		First(&contractID).Error
	return
}

func (mod *Module) SaveSignedNodeContract(c *user.SignedNodeContract) (err error) {
	contractID, err := astral.ResolveObjectID(c)
	if err != nil {
		return
	}

	// check if already saved
	if mod.ContractExists(contractID) {
		return nil
	}

	if c.IsExpired() {
		return errors.New("contract expired")
	}

	if err = c.VerifySigs(); err != nil {
		return fmt.Errorf("verify: %v", err)
	}

	return mod.db.Create(&dbNodeContract{
		ObjectID:  contractID,
		UserID:    c.UserID,
		NodeID:    c.NodeID,
		ExpiresAt: c.ExpiresAt.Time().UTC(),
	}).Error
}

func (mod *Module) LocalContract() (c *user.SignedNodeContract, err error) {
	if mod.userID.IsZero() {
		return nil, errors.New("local user not set")
	}
	var cid object.ID

	// first try loading an existing contract
	if cid, err = mod.findContractID(mod.userID, mod.node.Identity()); err == nil {
		c, err = objects.Load[*user.SignedNodeContract](context.Background(), mod.Objects, cid, astral.DefaultScope())
		if err == nil {
			return
		}
	}

	// then create and sign a new contract
	c = &user.SignedNodeContract{
		NodeContract: &user.NodeContract{
			UserID:    mod.UserID(),
			NodeID:    mod.node.Identity(),
			ExpiresAt: astral.Time(time.Now().Add(defaultContractValidity).UTC()),
		},
	}

	// sign with node key
	c.NodeSig, err = mod.Keys.Sign(c.NodeID, c.Hash())
	if err != nil {
		return
	}

	// sign with user key
	c.UserSig, err = mod.Keys.Sign(c.UserID, c.Hash())
	if err != nil {
		return
	}

	var b = &bytes.Buffer{}
	_, err = c.WriteTo(b)
	if err != nil {
		return
	}

	err = mod.SaveSignedNodeContract(c)
	if err != nil {
		return
	}

	_, err = mod.Objects.Store(c)
	if err != nil {
		return
	}

	err = mod.assets.Write(assetLocalContract, b.Bytes())
	return
}

func (mod *Module) String() string {
	return user.ModuleName
}
