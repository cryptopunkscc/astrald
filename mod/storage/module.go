package storage

import (
	"context"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/mod/admin/api"
	"github.com/cryptopunkscc/astrald/mod/sdp/api"
	storage "github.com/cryptopunkscc/astrald/mod/storage/api"
	"github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/node/events"
	"github.com/cryptopunkscc/astrald/tasks"
	"gorm.io/gorm"
)

var _ storage.API = &Module{}

type Module struct {
	node   node.Node
	config Config
	db     *gorm.DB
	log    *log.Logger
	events events.Queue
	ctx    context.Context
	sdp    sdp.API

	access *AccessManager
	data   *DataManager
}

func (mod *Module) Run(ctx context.Context) error {
	mod.ctx = ctx

	mod.sdp, _ = mod.node.Modules().Find("sdp").(sdp.API)

	// inject admin command
	if adm, _ := mod.node.Modules().Find("admin").(admin.API); adm != nil {
		adm.AddCommand("storage", NewAdmin(mod))
	}

	var runners = []tasks.Runner{
		NewReadService(mod),
	}

	tasks.Group(runners...).Run(ctx)

	<-ctx.Done()

	return nil
}

func (mod *Module) Access() storage.AccessManager {
	return mod.access
}

func (mod *Module) Data() storage.DataManager {
	return mod.data
}

func (mod *Module) Events() *events.Queue {
	return &mod.events
}
