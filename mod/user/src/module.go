package user

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/log"
	"github.com/cryptopunkscc/astrald/core/assets"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/mod/shell"
	"github.com/cryptopunkscc/astrald/mod/user"
	"github.com/cryptopunkscc/astrald/object"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
	"sync"
	"time"
)

const assetLocalContract = "mod.user.local_contract"
const defaultContractValidity = 365 * 24 * time.Hour

var _ user.Module = &Module{}

type Module struct {
	Deps
	config Config
	node   astral.Node
	log    *log.Logger
	assets assets.Assets
	db     *gorm.DB
	userID *astral.Identity
	user   *user.SignedNodeContract
	mu     sync.Mutex
	ops    shell.Scope
}

func (mod *Module) Run(ctx *astral.Context) error {
	<-ctx.Done()
	return nil
}

func (mod *Module) Nodes(userID *astral.Identity) (nodes []*astral.Identity) {
	err := mod.db.
		Model(&dbNodeContract{}).
		Where("expires_at > ?", time.Now().UTC()).
		Where("user_id = ?", userID).
		Where("node_id != ?", mod.node.Identity()).
		Distinct("node_id").
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
	mod.mu.Lock()
	defer mod.mu.Unlock()

	err := mod.setUserID(userID)
	if err != nil {
		return err
	}

	return mod.storeUserID(userID)
}

func (mod *Module) setUserID(userID *astral.Identity) error {
	if mod.userID.IsEqual(userID) {
		return nil
	}

	mod.userID = userID
	mod.user = nil

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

func (mod *Module) Remote(nodeID *astral.Identity, callerID *astral.Identity) (user.Consumer, error) {
	if callerID.IsZero() {
		callerID = mod.node.Identity()
	}

	return NewConsumer(mod, callerID, nodeID), nil
}

func (mod *Module) ContractExists(contractID object.ID) (b bool) {
	mod.db.
		Model(&dbNodeContract{}).
		Where("object_id = ?", contractID).
		Select("count(*) > 0").
		First(&b)
	return
}

func (mod *Module) Scope() *shell.Scope {
	return &mod.ops
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

	// verify user signature
	err = mod.Keys.VerifyASN1(c.UserID, c.Hash(), c.UserSig)
	if err != nil {
		return fmt.Errorf("user signature: %v", err)
	}

	// verify node signature
	err = mod.Keys.VerifyASN1(c.NodeID, c.Hash(), c.NodeSig)
	if err != nil {
		return fmt.Errorf("node signature: %v", err)
	}

	mod.Objects.Store(c)

	return mod.db.Create(&dbNodeContract{
		ObjectID:  contractID,
		UserID:    c.UserID,
		NodeID:    c.NodeID,
		ExpiresAt: c.ExpiresAt.Time().UTC(),
	}).Error
}

func (mod *Module) LocalContract() (c *user.SignedNodeContract, err error) {
	mod.mu.Lock()
	defer mod.mu.Unlock()

	if mod.user != nil {
		if !mod.user.IsExpired() {
			return mod.user, nil
		}

		mod.log.Log("user contract has expired")
		mod.user = nil
	}

	if mod.userID.IsZero() {
		return nil, errors.New("local user not set")
	}
	var cid object.ID

	// first try loading an existing contract
	if cid, err = mod.findContractID(mod.userID, mod.node.Identity()); err == nil {
		mod.user, err = objects.Load[*user.SignedNodeContract](context.Background(), mod.Objects, cid, astral.DefaultScope())
		if err == nil {
			mod.log.Infov(1, "loaded user contract %v", cid)
			return mod.user, nil
		}
		mod.log.Errorv(2, "error loading contract %v: %v", cid, err)
	}

	mod.log.Info("signing new user contract since no local contract was found (%v)", err)

	// then create and sign a new contract
	c = &user.SignedNodeContract{
		NodeContract: &user.NodeContract{
			UserID:    mod.UserID(),
			NodeID:    mod.node.Identity(),
			ExpiresAt: astral.Time(time.Now().Add(defaultContractValidity).UTC()),
		},
	}

	// sign with node key
	c.NodeSig, err = mod.Keys.SignASN1(c.NodeID, c.Hash())
	if err != nil {
		return
	}

	// sign with user key
	c.UserSig, err = mod.Keys.SignASN1(c.UserID, c.Hash())
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
