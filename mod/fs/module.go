package fs

import (
	"context"
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

const nameReadOnly = "mod.fs.readonly"
const nameReadWrite = "mod.fs.readwrite"

var _ fs.API = &Module{}

type Module struct {
	config Config
	node   node.Node
	log    *log.Logger
	events events.Queue
	db     *gorm.DB
	ctx    context.Context

	storage storage.API
	index   *IndexService
	store   *StoreService
}

func (mod *Module) Prepare(ctx context.Context) error {
	var err error

	// set up dependencies
	mod.storage, err = storage.Load(mod.node)
	if err != nil {
		return err
	}

	// read only
	mod.storage.Data().AddReader(nameReadOnly, mod.index)
	mod.storage.Data().AddIndex(nameReadOnly, mod.index)

	// read write
	mod.storage.Data().AddReader(nameReadWrite, mod.store)
	mod.storage.Data().AddStore(nameReadWrite, mod.store)
	mod.storage.Data().AddIndex(nameReadWrite, mod.store)

	// inject admin command
	if adm, err := admin.Load(mod.node); err == nil {
		adm.AddCommand(fs.ModuleName, NewAdmin(mod))
	}

	return nil
}

func (mod *Module) Run(ctx context.Context) error {
	mod.ctx = ctx

	tasks.Group(mod.index).Run(ctx)

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
