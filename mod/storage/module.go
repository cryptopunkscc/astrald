package storage

import (
	"context"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/tasks"
	"gorm.io/gorm"
	"sync"
)

type Module struct {
	node    node.Node
	config  Config
	db      *gorm.DB
	sources map[*Source]struct{}
	log     *log.Logger
	mu      sync.Mutex
}

type Source struct {
	Service  string
	Identity id.Identity
}

func (mod *Module) Run(ctx context.Context) error {
	return tasks.Group(
		&RegisterService{Module: mod},
		&ReadService{Module: mod},
	).Run(ctx)
}

func (mod *Module) AddSource(source *Source) {
	mod.mu.Lock()
	defer mod.mu.Unlock()

	mod.sources[source] = struct{}{}
	mod.log.Info("%s registered source %s", source.Identity, source.Service)
}

func (mod *Module) RemoveSource(source *Source) {
	mod.mu.Lock()
	defer mod.mu.Unlock()

	delete(mod.sources, source)
	mod.log.Logv(1, "%s unregistered source %s", source.Identity, source.Service)
}
