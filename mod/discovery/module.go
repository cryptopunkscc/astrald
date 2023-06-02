package discovery

import (
	"context"
	"errors"
	"fmt"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/cslq"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/mod/discovery/proto"
	"github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/node/event"
	"github.com/cryptopunkscc/astrald/tasks"
	"reflect"
	"sync"
)

type Module struct {
	node      node.Node
	events    event.Queue
	config    Config
	log       *log.Logger
	sources   map[Source]id.Identity
	sourcesMu sync.Mutex
	cache     map[string][]proto.ServiceEntry
	cacheMu   sync.Mutex
}

func (m *Module) Run(ctx context.Context) error {
	return tasks.Group(
		&DiscoveryService{Module: m},
		&RegisterService{Module: m},
		&EventHandler{Module: m},
	).Run(ctx)
}

func (m *Module) AddSource(source Source, identity id.Identity) {
	m.sourcesMu.Lock()
	defer m.sourcesMu.Unlock()

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
	m.sourcesMu.Lock()
	defer m.sourcesMu.Unlock()

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

func (m *Module) QueryLocal(ctx context.Context, caller id.Identity, origin string) ([]proto.ServiceEntry, error) {
	var list = make([]proto.ServiceEntry, 0)

	var wg sync.WaitGroup

	for source := range m.sources {
		source := source

		wg.Add(1)
		go func() {
			defer wg.Done()

			slist, err := source.Discover(ctx, caller, origin)
			if err != nil {
				return
			}

			list = append(list, slist...)
		}()
	}

	wg.Wait()

	return list, nil
}

func (m *Module) QueryRemoteAs(ctx context.Context, remoteID id.Identity, callerID id.Identity) ([]proto.ServiceEntry, error) {
	if callerID.IsZero() {
		callerID = m.node.Identity()
	}

	if !callerID.IsEqual(m.node.Identity()) {
		return nil, errors.New("switching identities unimplemented")
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	q, err := m.node.Network().Query(ctx, remoteID, "services.discover")
	if err != nil {
		return nil, err
	}

	var list = make([]proto.ServiceEntry, 0)

	go func() {
		<-ctx.Done()
		q.Close()
	}()

	for err == nil {
		err = cslq.Invoke(q, func(msg proto.ServiceEntry) error {
			list = append(list, msg)
			return nil
		})
	}

	return list, nil
}

func (m *Module) setCache(identity id.Identity, list []proto.ServiceEntry) {
	m.cacheMu.Lock()
	defer m.cacheMu.Unlock()

	m.cache[identity.String()] = list
}

func (m *Module) getCache(identity id.Identity) []proto.ServiceEntry {
	m.cacheMu.Lock()
	defer m.cacheMu.Unlock()

	return m.cache[identity.String()]
}
