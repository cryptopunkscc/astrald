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

func (m *Module) Run(ctx context.Context) error {
	return tasks.Group(
		&RegisterService{Module: m},
		&ReadService{Module: m},
	).Run(ctx)
}

func (m *Module) AddSource(source *Source) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.sources[source] = struct{}{}
	m.log.Info("%s registered source %s", source.Identity, source.Service)
}

func (m *Module) RemoveSource(source *Source) {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.sources, source)
	m.log.Logv(1, "%s unregistered source %s", source.Identity, source.Service)
}
