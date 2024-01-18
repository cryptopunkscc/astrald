package acl

import (
	"context"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/mod/acl"
	"github.com/cryptopunkscc/astrald/mod/admin"
	"github.com/cryptopunkscc/astrald/mod/storage"
	"github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/node/assets"
	"github.com/cryptopunkscc/astrald/node/modules"
	"gorm.io/gorm"
)

var _ acl.Module = &Module{}

type Module struct {
	config  Config
	node    node.Node
	log     *log.Logger
	assets  assets.Assets
	storage storage.Module
	db      *gorm.DB
}

func (mod *Module) Prepare(ctx context.Context) error {
	if adm, err := modules.Load[admin.Module](mod.node, admin.ModuleName); err == nil {
		adm.AddCommand(acl.ModuleName, NewAdmin(mod))
	}

	mod.storage.Access().AddAccessVerifier(mod)

	return nil
}

func (mod *Module) Run(ctx context.Context) error {
	<-ctx.Done()

	return nil
}
