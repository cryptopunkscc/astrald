package data

import (
	"context"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/mod/admin/api"
	data "github.com/cryptopunkscc/astrald/mod/data/api"
	fs "github.com/cryptopunkscc/astrald/mod/fs/api"
	storage "github.com/cryptopunkscc/astrald/mod/storage/api"
	"github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/node/events"
	"github.com/cryptopunkscc/astrald/tasks"
	"gorm.io/gorm"
)

var _ data.API = &Module{}

type Module struct {
	node   node.Node
	config Config
	log    *log.Logger
	ctx    context.Context
	events events.Queue

	storage storage.API
	fs      fs.API
	db      *gorm.DB
}

func (mod *Module) Events() *events.Queue {
	return &mod.events
}

func (mod *Module) Prepare(ctx context.Context) error {
	var err error

	// set up dependencies
	mod.storage, err = storage.Load(mod.node)
	if err != nil {
		return err
	}

	mod.fs, _ = fs.Load(mod.node)

	// inject admin command
	if adm, err := admin.Load(mod.node); err == nil {
		adm.AddCommand(data.ModuleName, NewAdmin(mod))
	}

	return nil
}

func (mod *Module) Run(ctx context.Context) error {
	mod.ctx = ctx

	return tasks.Group(&IndexerService{Module: mod}).Run(ctx)
}
