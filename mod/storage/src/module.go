package storage

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/data"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/mod/storage"
	"github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/node/events"
	"github.com/cryptopunkscc/astrald/sig"
	"gorm.io/gorm"
	"io"
)

var _ storage.Module = &Module{}

// ReadAllMaxSize is the limit on data size accepted by ReadAll() (to avoid accidental OOM)
var ReadAllMaxSize uint64 = 1024 * 1024 * 1024

type Module struct {
	node   node.Node
	config Config
	db     *gorm.DB
	log    *log.Logger
	events events.Queue
	ctx    context.Context

	openers  sig.Map[string, *Opener]
	creators sig.Map[string, *Creator]
	purgers  sig.Map[string, storage.Purger]
}

func (mod *Module) Run(ctx context.Context) error {
	mod.ctx = ctx

	<-ctx.Done()

	return nil
}

func (mod *Module) ReadAll(id data.ID, opts *storage.OpenOpts) ([]byte, error) {
	if id.Size > ReadAllMaxSize {
		return nil, errors.New("data too big")
	}
	r, err := mod.Open(id, opts)
	if err != nil {
		return nil, err
	}

	return io.ReadAll(r)
}

func (mod *Module) Put(bytes []byte, opts *storage.CreateOpts) (data.ID, error) {
	if opts == nil {
		opts = &storage.CreateOpts{Alloc: len(bytes)}
	}

	w, err := mod.Create(opts)
	if err != nil {
		return data.ID{}, err
	}
	defer w.Discard()

	_, err = w.Write(bytes)
	if err != nil {
		return data.ID{}, err
	}

	return w.Commit()
}

func (mod *Module) Events() *events.Queue {
	return &mod.events
}
