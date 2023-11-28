package sdp

import (
	"context"
	"fmt"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/cslq"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/mod/admin"
	"github.com/cryptopunkscc/astrald/mod/router"
	"github.com/cryptopunkscc/astrald/mod/sdp/proto"
	"github.com/cryptopunkscc/astrald/net"
	"github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/node/assets"
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
	assets    assets.Store
	log       *log.Logger
	sources   map[Source]struct{}
	sourcesMu sync.Mutex
	cache     map[string][]ServiceEntry
	routerMod *router.Module
	cacheMu   sync.Mutex
	ctx       context.Context
}

func (m *Module) Run(ctx context.Context) error {
	m.ctx = ctx

	m.routerMod, _ = modules.Find[*router.Module](m.node.Modules())

	// inject admin command
	if adm, err := modules.Find[*admin.Module](m.node.Modules()); err == nil {
		adm.AddCommand(ModuleName, NewAdmin(m))
	}

	return tasks.Group(
		&DiscoveryService{Module: m},
		&EventHandler{Module: m},
	).Run(ctx)
}

func (m *Module) AddSource(source Source) {
	m.sourcesMu.Lock()
	defer m.sourcesMu.Unlock()

	m.sources[source] = struct{}{}

	var s string
	if stringer, ok := source.(fmt.Stringer); ok {
		s = stringer.String()
	} else {
		s = reflect.TypeOf(source).String()
	}
	m.log.Infov(1, "registered source: %s", s)
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

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	if callerID.PrivateKey() == nil {
		keyStore, err := m.assets.KeyStore()
		if err != nil {
			return nil, err
		}

		callerID, err = keyStore.Find(callerID)
		if err != nil {
			return nil, err
		}
	}

	q, err := net.Route(ctx,
		m.node.Router(),
		net.NewQuery(callerID, remoteID, DiscoverServiceName),
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
		err = cslq.Invoke(q, func(msg proto.ServiceEntry) error {
			list = append(list, ServiceEntry(msg))
			if !msg.Identity.IsEqual(remoteID) {
				if m.routerMod != nil {
					m.routerMod.SetRouter(msg.Identity, remoteID)
				}
			}
			return nil
		})
	}

	return list, nil
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
