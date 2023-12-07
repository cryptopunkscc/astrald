package acl

import (
	"context"
	"github.com/cryptopunkscc/astrald/log"
	acl "github.com/cryptopunkscc/astrald/mod/acl/api"
	admin "github.com/cryptopunkscc/astrald/mod/admin/api"
	storage "github.com/cryptopunkscc/astrald/mod/storage/api"
	"github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/node/assets"
	"gorm.io/gorm"
)

var _ acl.API = &Module{}

type Module struct {
	config  Config
	node    node.Node
	log     *log.Logger
	assets  assets.Store
	storage storage.API
	db      *gorm.DB
}

func (mod *Module) Prepare(ctx context.Context) error {
	mod.storage, _ = storage.Load(mod.node)

	if m, err := admin.Load(mod.node); err == nil {
		m.AddCommand(acl.ModuleName, NewAdmin(mod))
	}

	mod.storage.Access().AddAccessVerifier(mod)

	return nil
}

func (mod *Module) Run(ctx context.Context) error {
	<-ctx.Done()
	return nil
}
