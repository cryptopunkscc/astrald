package storage

import (
	"context"
	"github.com/cryptopunkscc/astrald/data"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/mod/admin/api"
	"github.com/cryptopunkscc/astrald/mod/sdp/api"
	storage "github.com/cryptopunkscc/astrald/mod/storage/api"
	"github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/node/events"
	"github.com/cryptopunkscc/astrald/tasks"
	"gorm.io/gorm"
	"io"
	"sync"
)

var _ storage.API = &Module{}

type Module struct {
	node   node.Node
	config Config
	db     *gorm.DB
	log    *log.Logger
	events events.Queue
	ctx    context.Context
	sdp    sdp.API

	localFiles *LocalFiles

	readers   []storage.Reader
	readersMu sync.Mutex

	accessCheckers   map[AccessChecker]struct{}
	accessCheckersMu sync.Mutex
}

func (mod *Module) Run(ctx context.Context) error {
	mod.ctx = ctx

	mod.sdp, _ = mod.node.Modules().Find("sdp").(sdp.API)

	// inject admin command
	if adm, _ := mod.node.Modules().Find("admin").(admin.API); adm != nil {
		adm.AddCommand("storage", NewAdmin(mod))
	}

	var runners = []tasks.Runner{
		mod.localFiles,
		NewReadService(mod),
	}

	tasks.Group(runners...).Run(ctx)

	<-ctx.Done()

	return nil
}

func (mod *Module) Read(id data.ID, offset int, length int) (io.ReadCloser, error) {
	mod.readersMu.Lock()
	defer mod.readersMu.Unlock()

	for _, source := range mod.readers {
		r, err := source.Read(id, offset, length)
		if err == nil {
			return r, nil
		}
	}

	return nil, storage.ErrNotFound
}

func (mod *Module) LocalFiles() storage.LocalFiles {
	return mod.localFiles
}
