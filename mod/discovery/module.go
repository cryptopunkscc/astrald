package discovery

import (
	"context"
	"errors"
	"fmt"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/cslq"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/mod/admin"
	"github.com/cryptopunkscc/astrald/mod/discovery/rpc"
	"github.com/cryptopunkscc/astrald/mod/router"
	"github.com/cryptopunkscc/astrald/net"
	"github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/node/events"
	"github.com/cryptopunkscc/astrald/node/modules"
	"github.com/cryptopunkscc/astrald/tasks"
	"reflect"
	"sync"
)

type Module struct {
	node      node.Node
	events    events.Queue
	config    Config
	log       *log.Logger
	sources   map[Source]id.Identity
	sourcesMu sync.Mutex
	cache     map[string][]ServiceEntry
	routes    map[string]id.Identity
	cacheMu   sync.Mutex
	ctx       context.Context
}

func (m *Module) Run(ctx context.Context) error {
	m.ctx = ctx

	// inject admin command
	if adm, err := modules.Find[*admin.Module](m.node.Modules()); err == nil {
		adm.AddCommand("discovery", NewAdmin(m))
	}

	// register as a router
	if coreRouter, ok := m.node.Router().(*node.CoreRouter); ok {
		coreRouter.Routers.AddRouter(m)
	}

	return tasks.Group(
		&DiscoveryService{Module: m},
		&RegisterService{Module: m},
		&EventHandler{Module: m},
	).Run(ctx)
}

func (m *Module) AddSourceContext(ctx context.Context, source Source, identity id.Identity) {
	m.AddSource(source, identity)
	go func() {
		<-ctx.Done()
		m.RemoveSource(source)
	}()
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
	m.log.Infov(1, "registered source: %s (%s)", s, identity)
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

func (m *Module) QueryLocal(ctx context.Context, caller id.Identity, origin string) ([]ServiceEntry, error) {
	var list = make([]ServiceEntry, 0)

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

func (m *Module) QueryRemoteAs(ctx context.Context, remoteID id.Identity, callerID id.Identity) ([]ServiceEntry, error) {
	if callerID.IsZero() {
		callerID = m.node.Identity()
	}

	if !callerID.IsEqual(m.node.Identity()) {
		return nil, errors.New("switching identities unimplemented")
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	q, err := net.Route(ctx,
		m.node.Network(),
		net.NewQuery(m.node.Identity(), remoteID, discoverServiceName),
	)
	if err != nil {
		return nil, err
	}

	var list = make([]ServiceEntry, 0)

	go func() {
		<-ctx.Done()
		q.Close()
	}()

	for err == nil {
		err = cslq.Invoke(q, func(msg rpc.ServiceEntry) error {
			list = append(list, ServiceEntry(msg))
			if !msg.Identity.IsEqual(remoteID) {
				m.routes[msg.Identity.PublicKeyHex()] = remoteID
			}
			return nil
		})
	}

	return list, nil
}

func (m *Module) RouteQuery(ctx context.Context, query net.Query, caller net.SecureWriteCloser, hints net.Hints) (net.SecureWriteCloser, error) {
	relay, found := m.routes[query.Target().PublicKeyHex()]
	if !found {
		return net.RouteNotFound(m)
	}

	routeMod, err := modules.Find[*router.Module](m.node.Modules())
	if err != nil {
		return nil, err
	}

	return routeMod.RouteVia(ctx, relay, query, caller, hints)
}

func (m *Module) setCache(identity id.Identity, list []ServiceEntry) {
	m.cacheMu.Lock()
	defer m.cacheMu.Unlock()

	m.cache[identity.String()] = list
}

func (m *Module) getCache(identity id.Identity) []ServiceEntry {
	m.cacheMu.Lock()
	defer m.cacheMu.Unlock()

	return m.cache[identity.String()]
}
