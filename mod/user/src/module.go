package user

import (
	"context"
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
	"github.com/cryptopunkscc/astrald/mod/relay"
	"github.com/cryptopunkscc/astrald/mod/sets"
	"github.com/cryptopunkscc/astrald/mod/shares"
	"github.com/cryptopunkscc/astrald/mod/user"
	"github.com/cryptopunkscc/astrald/object"
	"gorm.io/gorm"
	"time"
)

var _ user.Module = &Module{}

type Module struct {
	config  Config
	node    astral.Node
	log     *log.Logger
	assets  assets.Assets
	db      *gorm.DB
	objects objects.Module
	shares  shares.Module
	content content.Module
	relay   relay.Module
	keys    keys.Module
	admin   admin.Module
	apphost apphost.Module
	sets    sets.Module
	dir     dir.Module
	auth    auth.Module

	*routers.PathRouter
	userID         id.Identity
	userCert       []byte
	profileService *ProfileService
}

func (mod *Module) Run(ctx context.Context) error {
	go mod.rescanContracts(ctx)

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
		Where("expires_at > ?", time.Now()).
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

func (mod *Module) rescanContracts(ctx context.Context) error {
	opts := &content.ScanOpts{
		Type: (&user.NodeContract{}).ObjectType(),
	}

	for info := range mod.content.Scan(ctx, opts) {
		contract, err := objects.Load[*user.NodeContract](ctx, mod.objects, info.ObjectID, astral.DefaultScope())
		if err != nil {
			continue
		}

		mod.setCache(info.ObjectID, contract)
	}
	return nil
}

func (mod *Module) setUserID(userID id.Identity) error {
	cert, err := mod.loadCert(userID, mod.node.Identity(), true)
	if err != nil {
		return err
	}

	mod.userID = userID
	mod.userCert = cert

	mod.log.Info("user identity set to %v", mod.userID)

	return nil
}

func (mod *Module) setCache(objectID object.ID, contract *user.NodeContract) error {
	if err := contract.Validate(); err != nil {
		return err
	}
	return mod.db.Create(&dbNodeContract{
		ObjectID:  objectID,
		UserID:    contract.UserID,
		NodeID:    contract.NodeID,
		ExpiresAt: contract.ExpiresAt,
	}).Error
}

func (mod *Module) clearCache(objectID object.ID) error {
	return mod.db.Where("object_id = ?", objectID).Delete(&dbNodeContract{}).Error
}

func (mod *Module) getCache(objectID object.ID) *user.NodeContract {
	var row dbNodeContract
	err := mod.db.First(&row, "object_id = ?", objectID).Error
	if err != nil {
		return nil
	}
	return &user.NodeContract{
		UserID:    row.UserID,
		NodeID:    row.NodeID,
		ExpiresAt: row.ExpiresAt,
	}
}
