package fs

import (
	"context"
	"github.com/cryptopunkscc/astrald/data"
	"github.com/cryptopunkscc/astrald/log"
	_data "github.com/cryptopunkscc/astrald/mod/data"
	"github.com/cryptopunkscc/astrald/mod/fs"
	"github.com/cryptopunkscc/astrald/mod/storage"
	"github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/node/events"
	"github.com/cryptopunkscc/astrald/tasks"
	"gorm.io/gorm"
)

const nameReadOnly = "mod.fs.readonly"
const nameReadWrite = "mod.fs.readwrite"

var _ fs.Module = &Module{}

type Module struct {
	config Config
	node   node.Node
	log    *log.Logger
	events events.Queue
	db     *gorm.DB
	ctx    context.Context

	storage storage.Module
	data    _data.Module
	index   *IndexService
	store   *StoreService
}

func (mod *Module) DescribeData(ctx context.Context, dataID data.ID, opts *_data.DescribeOpts) []_data.Descriptor {
	var desc fs.FileDescriptor
	var files = mod.dbFindByID(dataID)

	if len(files) == 0 {
		return nil
	}

	for _, file := range files {
		desc.Paths = append(desc.Paths, file.Path)
	}

	return []_data.Descriptor{
		{
			Type: fs.FileDescriptorType,
			Data: desc,
		},
	}
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
