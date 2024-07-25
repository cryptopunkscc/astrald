package gateway

import (
	"context"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/core"
	"github.com/cryptopunkscc/astrald/id"
	"github.com/cryptopunkscc/astrald/lib/routers"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/mod/dir"
	"github.com/cryptopunkscc/astrald/mod/exonet"
	"github.com/cryptopunkscc/astrald/mod/nodes"
	"github.com/cryptopunkscc/astrald/tasks"
	"sync"
)

const NetworkName = "gw"

type Module struct {
	*routers.PathRouter
	config      Config
	node        astral.Node
	log         *log.Logger
	ctx         context.Context
	dialer      *Dialer
	subscribers map[string]*Subscriber
	mu          sync.Mutex
	nodes       nodes.Module
	exonet      exonet.Module
	dir         dir.Module
}

func (mod *Module) Prepare(ctx context.Context) (err error) {
	mod.dir, err = core.Load[dir.Module](mod.node, dir.ModuleName)
	if err != nil {
		return
	}

	mod.nodes, err = core.Load[nodes.Module](mod.node, nodes.ModuleName)
	if err != nil {
		return
	}

	mod.exonet, err = core.Load[exonet.Module](mod.node, exonet.ModuleName)
	if err != nil {
		return
	}

	mod.exonet.SetDialer("gw", mod.dialer)
	mod.exonet.SetUnpacker("gw", mod)
	mod.exonet.SetParser("gw", mod)

	return
}

func (mod *Module) Run(ctx context.Context) error {
	mod.ctx = ctx

	for _, gateName := range mod.config.Subscribe {
		var gateID id.Identity

		if info, err := mod.nodes.ParseInfo(gateName); err == nil {
			err = mod.nodes.AddEndpoint(info.Identity, info.Endpoints...)
			if err != nil {
				mod.log.Error("config error: endpoints: %v", err)
				continue
			}
			err = mod.dir.SetAlias(info.Identity, info.Alias)
			if err != nil {
				mod.log.Error("config error: set alias: %v", err)
				continue
			}
			gateID = info.Identity
		} else {
			gateID, err = mod.dir.Resolve(gateName)
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
		&RouteService{Module: mod, router: mod.node.Router()},
	).Run(ctx)
}

func (mod *Module) Subscribe(gateway id.Identity) error {
	mod.mu.Lock()
	defer mod.mu.Unlock()

	if gateway.IsEqual(mod.node.Identity()) {
		return ErrSelfGateway
	}

	var hex = gateway.PublicKeyHex()

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

func (mod *Module) Unsubscribe(gateway id.Identity) error {
	mod.mu.Lock()
	defer mod.mu.Unlock()

	s, found := mod.subscribers[gateway.PublicKeyHex()]
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
