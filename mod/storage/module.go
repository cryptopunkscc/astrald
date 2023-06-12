package storage

import (
	"context"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/mod/admin"
	"github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/node/events"
	"github.com/cryptopunkscc/astrald/node/modules"
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

	dataSources   map[*DataSource]struct{}
	dataSourcesMu sync.Mutex

	accessCheckers   map[AccessChecker]struct{}
	accessCheckersMu sync.Mutex
}

func (mod *Module) Run(ctx context.Context) error {
	// inject admin command
	if adm, err := modules.Find[*admin.Module](mod.node.Modules()); err == nil {
		adm.AddCommand("storage", NewAdmin(mod))
	}

	return tasks.Group(
		&RegisterService{Module: mod},
		&ReadService{Module: mod},
	).Run(ctx)
}
