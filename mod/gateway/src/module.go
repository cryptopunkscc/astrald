package gateway

import (
	"context"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/lib/routers"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/mod/dir"
	"github.com/cryptopunkscc/astrald/mod/exonet"
	"github.com/cryptopunkscc/astrald/mod/nodes"
	"github.com/cryptopunkscc/astrald/tasks"
	"sync"
)

const NetworkName = "gw"

type Deps struct {
	Dir    dir.Module
	Exonet exonet.Module
	Nodes  nodes.Module
}

type Module struct {
	Deps
	*routers.PathRouter
	config      Config
	node        astral.Node
	log         *log.Logger
	ctx         context.Context
	dialer      *Dialer
	subscribers map[string]*Subscriber
	mu          sync.Mutex
}

func (mod *Module) Run(ctx context.Context) error {
	mod.ctx = ctx

	for _, gateName := range mod.config.Subscribe {
		var gateID *astral.Identity

		if info, err := mod.Nodes.ParseInfo(gateName); err == nil {
			err = mod.Nodes.AddEndpoint(info.Identity, info.Endpoints...)
			if err != nil {
				mod.log.Error("config error: endpoints: %v", err)
				continue
			}
			err = mod.Dir.SetAlias(info.Identity, info.Alias)
			if err != nil {
				mod.log.Error("config error: set alias: %v", err)
				continue
			}
			gateID = info.Identity
		} else {
			gateID, err = mod.Dir.ResolveIdentity(gateName)
			if err != nil {
				mod.log.Error("config error: cannot resolve %s: %v", gateName, err)
				continue
			}
		}

		if !gateID.IsZero() {
			mod.Subscribe(gateID)
		}
	}

	return tasks.Group(
		&SubscribeService{Module: mod},
		&RouteService{Module: mod, router: mod.node},
	).Run(ctx)
}

func (mod *Module) Subscribe(gateway *astral.Identity) error {
	mod.mu.Lock()
	defer mod.mu.Unlock()

	if gateway.IsEqual(mod.node.Identity()) {
		return ErrSelfGateway
	}

	var hex = gateway.String()

	if _, found := mod.subscribers[hex]; found {
		return ErrAlreadySubscribed
	}

	var s = NewSubscriber(gateway, mod.node, mod.log)
	mod.subscribers[hex] = s

	go func() {
		err := s.Run(mod.ctx)
		if err != nil {
			mod.log.Errorv(1, "gateway %v subscriber ended with error: %v", gateway, err)
		}
		mod.mu.Lock()
		defer mod.mu.Unlock()

		delete(mod.subscribers, hex)
	}()

	return nil
}

func (mod *Module) Unsubscribe(gateway *astral.Identity) error {
	mod.mu.Lock()
	defer mod.mu.Unlock()

	s, found := mod.subscribers[gateway.String()]
	if !found {
		return ErrNotSubscribed
	}

	s.Cancel()
	return nil
}

func (mod *Module) Endpoints() []exonet.Endpoint {
	var list = make([]exonet.Endpoint, 0)

	for _, s := range mod.subscribers {
		list = append(list, NewEndpoint(s.Gateway(), mod.node.Identity()))
	}

	return list
}

func (mod *Module) String() string {
	return ModuleName
}
