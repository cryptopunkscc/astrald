package media

import (
	"context"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/mod/content"
	"github.com/cryptopunkscc/astrald/mod/sets"
	"github.com/cryptopunkscc/astrald/mod/storage"
	"github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/resources"
	"github.com/cryptopunkscc/astrald/tasks"
	"gorm.io/gorm"
)

type Module struct {
	config Config
	node   node.Node
	db     *gorm.DB
	log    *log.Logger
	assets resources.Resources

	content content.Module
	storage storage.Module
	sets    sets.Module

	indexer *IndexerService
}

func (mod *Module) Run(ctx context.Context) error {
	return tasks.Group(
		mod.indexer,
	).Run(ctx)
}
