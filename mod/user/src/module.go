package user

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/core/assets"
	"github.com/cryptopunkscc/astrald/id"
	"github.com/cryptopunkscc/astrald/lib/routers"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/mod/admin"
	"github.com/cryptopunkscc/astrald/mod/apphost"
	"github.com/cryptopunkscc/astrald/mod/auth"
	"github.com/cryptopunkscc/astrald/mod/content"
	"github.com/cryptopunkscc/astrald/mod/dir"
	"github.com/cryptopunkscc/astrald/mod/keys"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/mod/sets"
	"github.com/cryptopunkscc/astrald/mod/user"
	"github.com/cryptopunkscc/astrald/object"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"time"
)

const assetLocalContract = "mod.user.local_contract"

var _ user.Module = &Module{}

type Deps struct {
	Admin   admin.Module
	Apphost apphost.Module
	Auth    auth.Module
	Content content.Module
	Dir     dir.Module
	Objects objects.Module
	Keys    keys.Module
	Sets    sets.Module
}

type Module struct {
	Deps
	config Config
	node   astral.Node
	log    *log.Logger
	assets assets.Assets
	db     *gorm.DB

	*routers.PathRouter
	userID         id.Identity
	userCert       []byte
	profileService *ProfileService
}

func (mod *Module) Run(ctx context.Context) error {
	<-ctx.Done()
	return nil
}

func (mod *Module) Nodes(userID id.Identity) (nodes []id.Identity) {
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

func (mod *Module) Owner(nodeID id.Identity) (userID id.Identity) {
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

func (mod *Module) UserID() id.Identity {
	return mod.userID
}

func (mod *Module) SetUserID(userID id.Identity) error {
	err := mod.setUserID(userID)
	if err != nil {
		return err
	}

	return mod.storeUserID(userID)
}

func (mod *Module) setUserID(userID id.Identity) error {
	mod.userID = userID

	mod.log.Info("user identity set to %v", mod.userID)

	return nil
}

func (mod *Module) AddContact(userID id.Identity) error {
	return mod.db.Clauses(clause.OnConflict{DoNothing: true}).Create(&dbContact{
		UserID: userID,
	}).Error
}

func (mod *Module) RemoveContact(userID id.Identity) error {
	return mod.db.Delete(&dbContact{UserID: userID}).Error
}

func (mod *Module) IsContact(userID id.Identity) (b bool) {
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

func (mod *Module) Contacts() (contacts []id.Identity) {
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

func (mod *Module) findContractID(userID id.Identity, nodeID id.Identity) (contractID object.ID, err error) {
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
		ExpiresAt: time.Time(c.ExpiresAt),
	}).Error
}

func (mod *Module) LocalContract() (c *user.SignedNodeContract, err error) {
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
			ExpiresAt: astral.Time(time.Now().Add(24 * time.Hour)),
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
