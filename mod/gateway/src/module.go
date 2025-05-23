package gateway

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/log"
	"github.com/cryptopunkscc/astrald/lib/routers"
	"github.com/cryptopunkscc/astrald/mod/dir"
	"github.com/cryptopunkscc/astrald/mod/exonet"
	gateway2 "github.com/cryptopunkscc/astrald/mod/gateway"
	"github.com/cryptopunkscc/astrald/mod/nodes"
	"github.com/cryptopunkscc/astrald/tasks"
	"strings"
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
	ctx         *astral.Context
	dialer      *Dialer
	subscribers map[string]*Subscriber
	mu          sync.Mutex
}

func (mod *Module) Run(ctx *astral.Context) error {
	mod.ctx = ctx.IncludeZone(astral.ZoneNetwork)

	mod.subscribeToGateways()

	return tasks.Group(
		&SubscribeService{Module: mod},
		&RouteService{Module: mod, router: mod.node},
	).Run(ctx)
}

func (mod *Module) subscribeToGateways() {
	for _, gateName := range mod.config.Subscribe {
		var gateID *astral.Identity

		if after, found := strings.CutPrefix(gateName, "node1"); found && len(after) > 32 {
			var info nodes.NodeInfo

			err := info.UnmarshalText([]byte(after))
			if err != nil {
				mod.log.Error("parse node info: %v", err)
				continue
			}

			// try to set alias
			err = mod.Dir.SetAlias(info.Identity, string(info.Alias))
			if err != nil {
				mod.log.Error("set alias: %v", err)
			}

			// save endpoints
			for _, ep := range info.Endpoints {
				err = mod.Nodes.AddEndpoint(info.Identity, ep)
				if err != nil {
					mod.log.Error("add endpoint: %v", err)
					continue
				}
			}

			// subscribe
			err = mod.Subscribe(info.Identity)
			if err != nil {
				mod.log.Error("subscribe: %v", err)
			}
			continue
		}

		gateID, err := mod.Dir.ResolveIdentity(gateName)
		if err != nil {
			mod.log.Error("resolve identity %v: %v", gateName, err)
			continue
		}

		err = mod.Subscribe(gateID)
		if err != nil {
			mod.log.Error("subscribe: %v", err)
		}
	}
}

func (mod *Module) Subscribe(gateway *astral.Identity) error {
	mod.mu.Lock()
	defer mod.mu.Unlock()

	switch {
	case gateway.IsZero():
		return ErrInvalidGateway
	case gateway.IsEqual(mod.node.Identity()):
		return ErrInvalidGateway
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
		list = append(list, gateway2.NewEndpoint(s.Gateway(), mod.node.Identity()))
	}

	return list
}

func (mod *Module) String() string {
	return ModuleName
}
