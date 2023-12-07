package sdp

import (
	"context"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/cslq"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/mod/admin/api"
	"github.com/cryptopunkscc/astrald/mod/router/api"
	. "github.com/cryptopunkscc/astrald/mod/sdp/api"
	"github.com/cryptopunkscc/astrald/mod/sdp/proto"
	"github.com/cryptopunkscc/astrald/net"
	"github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/node/assets"
	"github.com/cryptopunkscc/astrald/node/events"
	"github.com/cryptopunkscc/astrald/tasks"
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
	router    router.API
	cacheMu   sync.Mutex
	ctx       context.Context
}

func (mod *Module) Prepare(ctx context.Context) error {
	mod.router, _ = router.Load(mod.node)

	// inject admin command
	if adm, err := admin.Load(mod.node); err == nil {
		adm.AddCommand(ModuleName, NewAdmin(mod))
	}

	return nil
}

func (mod *Module) Run(ctx context.Context) error {
	mod.ctx = ctx

	return tasks.Group(
		&DiscoveryService{Module: mod},
		&EventHandler{Module: mod},
	).Run(ctx)
}

func (mod *Module) AddSource(source Source) {
	mod.sourcesMu.Lock()
	defer mod.sourcesMu.Unlock()

	mod.sources[source] = struct{}{}
}

func (mod *Module) RemoveSource(source Source) {
	mod.sourcesMu.Lock()
	defer mod.sourcesMu.Unlock()

	_, found := mod.sources[source]
	if !found {
		return
	}

	delete(mod.sources, source)
}

func (mod *Module) QueryLocal(ctx context.Context, caller id.Identity, origin string) ([]ServiceEntry, error) {
	var list = make([]ServiceEntry, 0)

	var wg sync.WaitGroup

	for source := range mod.sources {
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

func (mod *Module) QueryRemoteAs(ctx context.Context, remoteID id.Identity, callerID id.Identity) ([]ServiceEntry, error) {
	if callerID.IsZero() {
		callerID = mod.node.Identity()
	}

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	if callerID.PrivateKey() == nil {
		keyStore, err := mod.assets.KeyStore()
		if err != nil {
			return nil, err
		}

		callerID, err = keyStore.Find(callerID)
		if err != nil {
			return nil, err
		}
	}

	q, err := net.Route(ctx,
		mod.node.Router(),
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
				if mod.router != nil {
					mod.router.SetRouter(msg.Identity, remoteID)
				}
			}
			return nil
		})
	}

	return list, nil
}

func (mod *Module) setCache(identity id.Identity, list []ServiceEntry) {
	mod.cacheMu.Lock()
	defer mod.cacheMu.Unlock()

	mod.cache[identity.String()] = list
}

func (mod *Module) getCache(identity id.Identity) []ServiceEntry {
	mod.cacheMu.Lock()
	defer mod.cacheMu.Unlock()

	return mod.cache[identity.String()]
}
