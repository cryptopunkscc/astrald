package storage

import (
	"context"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/mod/admin/api"
	"github.com/cryptopunkscc/astrald/mod/sdp/api"
	"github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/node/events"
	"github.com/cryptopunkscc/astrald/tasks"
	"gorm.io/gorm"
	"sync"
)

type Module struct {
	node   node.Node
	config Config
	db     *gorm.DB
	log    *log.Logger
	events events.Queue
	ctx    context.Context
	sdp    sdp.API

	dataSources   map[*DataSource]struct{}
	dataSourcesMu sync.Mutex

	accessCheckers   map[AccessChecker]struct{}
	accessCheckersMu sync.Mutex
}

func (mod *Module) Run(ctx context.Context) error {
	mod.ctx = ctx

	mod.sdp, _ = mod.node.Modules().Find("sdp").(sdp.API)

	// inject admin command
	if adm, _ := mod.node.Modules().Find("admin").(admin.API); adm != nil {
		adm.AddCommand("storage", NewAdmin(mod))
	}

	return tasks.Group(
		&RegisterService{Module: mod},
		&ReadService{Module: mod},
	).Run(ctx)
}
