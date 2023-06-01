package discovery

import (
	"context"
	"fmt"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/mod/discovery/proto"
	"github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/node/event"
	"github.com/cryptopunkscc/astrald/tasks"
	"reflect"
	"sync"
)

type Module struct {
	node    node.Node
	events  event.Queue
	config  Config
	log     *log.Logger
	mu      sync.Mutex
	sources map[Source]id.Identity
}

func (m *Module) Run(ctx context.Context) error {
	return tasks.Group(
		&DiscoveryService{Module: m},
		&RegisterService{Module: m},
		&EventHandler{Module: m},
	).Run(ctx)
}

func (m *Module) AddSource(source Source, identity id.Identity) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.sources[source] = identity

	var s string
	if stringer, ok := source.(fmt.Stringer); ok {
		s = stringer.String()
	} else {
		s = reflect.TypeOf(source).String()
	}
	m.log.Info("registered source: %s (%s)", s, identity)
}

func (m *Module) RemoveSource(source Source) {
	m.mu.Lock()
	defer m.mu.Unlock()

	identity, found := m.sources[source]
	if !found {
		return
	}

	delete(m.sources, source)

	var s string
	if stringer, ok := source.(fmt.Stringer); ok {
		s = stringer.String()
	} else {
		s = reflect.TypeOf(source).String()
	}
	m.log.Logv(1, "unregistered source: %s (%s)", s, identity)
}

func (m *Module) DiscoverAs(ctx context.Context, identity id.Identity, origin string) ([]proto.ServiceEntry, error) {
	var list = make([]proto.ServiceEntry, 0)

	var wg sync.WaitGroup

	for source := range m.sources {
		source := source

		wg.Add(1)
		go func() {
			defer wg.Done()

			slist, err := source.Discover(ctx, identity, origin)
			if err != nil {
				return
			}

			list = append(list, slist...)
		}()
	}

	wg.Wait()

	return list, nil
}
