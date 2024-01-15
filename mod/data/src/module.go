package data

import (
	"context"
	"github.com/cryptopunkscc/astrald/cslq"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/mod/data"
	"github.com/cryptopunkscc/astrald/mod/fs"
	"github.com/cryptopunkscc/astrald/mod/storage"
	"github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/node/events"
	"github.com/cryptopunkscc/astrald/sig"
	"github.com/cryptopunkscc/astrald/tasks"
	"gorm.io/gorm"
)

var _ data.Module = &Module{}

type Module struct {
	node   node.Node
	config Config
	log    *log.Logger
	ctx    context.Context
	events events.Queue
	db     *gorm.DB

	describers sig.Set[data.Describer]
	storage    storage.Module
	fs         fs.Module
}

func (mod *Module) Run(ctx context.Context) error {
	mod.ctx = ctx

	return tasks.Group(&IndexService{Module: mod}).Run(ctx)
}

func (mod *Module) StoreADC0(t string, alloc int) (storage.DataWriter, error) {
	w, err := mod.storage.Data().Store(
		&storage.StoreOpts{
			Alloc: alloc + len(t) + 5,
		},
	)
	if err != nil {
		return nil, err
	}

	err = cslq.Encode(w, "v", data.ADC0Header(t))
	if err != nil {
		return nil, err
	}

	return w, nil
}

func (mod *Module) Events() *events.Queue {
	return &mod.events
}
