package data

import (
	"context"
	"errors"
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

func (mod *Module) Run(ctx context.Context) error {
	mod.ctx = ctx

	mod.storage, _ = mod.node.Modules().Find("storage").(storage.API)
	if mod.storage == nil {
		return errors.New("required storage module not loaded")
	}

	mod.fs, _ = mod.node.Modules().Find("fs").(fs.API)

	// inject admin command
	if adm, _ := mod.node.Modules().Find("admin").(admin.API); adm != nil {
		adm.AddCommand("data", NewAdmin(mod))
	}

	return tasks.Group(&IndexerService{Module: mod}).Run(ctx)
}
