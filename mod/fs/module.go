package fs

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/data"
	"github.com/cryptopunkscc/astrald/log"
	admin "github.com/cryptopunkscc/astrald/mod/admin/api"
	fs "github.com/cryptopunkscc/astrald/mod/fs/api"
	storage "github.com/cryptopunkscc/astrald/mod/storage/api"
	"github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/node/events"
	"github.com/cryptopunkscc/astrald/tasks"
	"gorm.io/gorm"
)

var _ fs.API = &Module{}

type Module struct {
	config Config
	node   node.Node
	log    *log.Logger
	events events.Queue
	db     *gorm.DB
	ctx    context.Context

	storage storage.API
	indexer *IndexerService
	storer  *StorerService
}

func (mod *Module) Run(ctx context.Context) error {
	mod.ctx = ctx

	// inject admin command
	if adm, _ := mod.node.Modules().Find("admin").(admin.API); adm != nil {
		adm.AddCommand("fs", NewAdmin(mod))
	}

	// set up dependencies
	mod.storage, _ = mod.node.Modules().Find("storage").(storage.API)
	if mod.storage == nil {
		return errors.New("storage module not found")
	}
	mod.storage.Data().AddReader(mod.indexer)
	mod.storage.Data().AddIndexer(mod.indexer)
	mod.storage.Data().AddReader(mod.storer)
	mod.storage.Data().AddIndexer(mod.storer)
	mod.storage.Data().AddStorer(mod.storer)

	tasks.Group(mod.indexer).Run(ctx)

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
