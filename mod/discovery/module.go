package discovery

import (
	"context"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/tasks"
	"sync"
)

type Module struct {
	node    node.Node
	config  Config
	log     *log.Logger
	mu      sync.Mutex
	sources map[*Source]struct{}
}

type Source struct {
	Identity id.Identity
	Service  string
}

func (m *Module) Run(ctx context.Context) error {
	return tasks.Group(
		&DiscoveryService{Module: m},
		&RegisterService{Module: m},
	).Run(ctx)
}

func (m *Module) AddSource(source *Source) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.sources[source] = struct{}{}
	m.log.Info("registered source: %s", source.Service)
}

func (m *Module) RemoveSource(source *Source) {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.sources, source)
	m.log.Logv(1, "unregistered source: %s", source.Service)
}
