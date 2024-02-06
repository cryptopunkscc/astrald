package fs

import (
	"context"
	"github.com/cryptopunkscc/astrald/data"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/mod/content"
	"github.com/cryptopunkscc/astrald/mod/fs"
	"github.com/cryptopunkscc/astrald/mod/sets"
	"github.com/cryptopunkscc/astrald/mod/storage"
	"github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/node/assets"
	"github.com/cryptopunkscc/astrald/node/events"
	"github.com/cryptopunkscc/astrald/tasks"
	"gorm.io/gorm"
)

var _ fs.Module = &Module{}

type Module struct {
	config Config
	node   node.Node
	assets assets.Assets
	log    *log.Logger
	events events.Queue
	db     *gorm.DB
	ctx    context.Context

	storage storage.Module
	content content.Module
	sets    sets.Module

	readonly  *IndexerService
	readwrite *ReadWriteService
	memStore  *MemStore
	memSet    sets.Set
	roSet     sets.Set
	rwSet     sets.Set
}

func (mod *Module) Prepare(ctx context.Context) error {
	return nil
}

func (mod *Module) Run(ctx context.Context) error {
	mod.ctx = ctx

	tasks.Group(mod.readonly).Run(ctx)

	<-ctx.Done()
	return nil
}

func (mod *Module) Find(dataID data.ID) []string {
	var list []string
	for _, row := range mod.dbFindByID(dataID) {
		list = append(list, row.Path)
	}
	return list
}
