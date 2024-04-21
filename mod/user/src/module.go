package user

import (
	"context"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/mod/admin"
	"github.com/cryptopunkscc/astrald/mod/apphost"
	"github.com/cryptopunkscc/astrald/mod/content"
	"github.com/cryptopunkscc/astrald/mod/dir"
	"github.com/cryptopunkscc/astrald/mod/discovery"
	"github.com/cryptopunkscc/astrald/mod/keys"
	"github.com/cryptopunkscc/astrald/mod/relay"
	"github.com/cryptopunkscc/astrald/mod/sets"
	"github.com/cryptopunkscc/astrald/mod/shares"
	"github.com/cryptopunkscc/astrald/mod/storage"
	"github.com/cryptopunkscc/astrald/mod/user"
	"github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/node/assets"
	"github.com/cryptopunkscc/astrald/node/router"
	"gorm.io/gorm"
)

var _ user.Module = &Module{}

type Module struct {
	config  Config
	node    node.Node
	log     *log.Logger
	assets  assets.Assets
	db      *gorm.DB
	storage storage.Module
	shares  shares.Module
	content content.Module
	sdp     discovery.Module
	relay   relay.Module
	keys    keys.Module
	admin   admin.Module
	apphost apphost.Module
	sets    sets.Module
	dir     dir.Module

	userID         id.Identity
	userCert       []byte
	routes         *router.PrefixRouter
	profileService *ProfileService
}

func (mod *Module) Run(ctx context.Context) error {
	<-ctx.Done()

	return nil
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
	cert, err := mod.loadCert(userID, mod.node.Identity(), true)
	if err != nil {
		return err
	}

	mod.userID = userID
	mod.userCert = cert

	mod.log.Info("user identity set to %v", mod.userID)

	return nil
}
