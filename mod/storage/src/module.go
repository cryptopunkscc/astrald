package storage

import (
	"context"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/mod/discovery"
	"github.com/cryptopunkscc/astrald/mod/storage"
	"github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/node/events"
	"github.com/cryptopunkscc/astrald/tasks"
	"gorm.io/gorm"
)

var _ storage.Module = &Module{}

type Module struct {
	node   node.Node
	config Config
	db     *gorm.DB
	log    *log.Logger
	events events.Queue
	ctx    context.Context
	sdp    discovery.Module

	access *AccessManager
	data   *DataManager
}

func (mod *Module) Run(ctx context.Context) error {
	mod.ctx = ctx

	tasks.Group(NewReadService(mod)).Run(ctx)

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
