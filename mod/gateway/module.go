package gateway

import (
	"context"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/mod/policy"
	"github.com/cryptopunkscc/astrald/net"
	"github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/node/modules"
	"github.com/cryptopunkscc/astrald/nodeinfo"
	"github.com/cryptopunkscc/astrald/tasks"
	"sync"
)

const NetworkName = "gw"

type Module struct {
	config      Config
	node        node.Node
	log         *log.Logger
	ctx         context.Context
	dialer      *Dialer
	subscribers map[string]*Subscriber
	mu          sync.Mutex
}

func (mod *Module) Run(ctx context.Context) error {
	mod.ctx = ctx

	for _, gateName := range mod.config.Subscribe {
		var gateID id.Identity

		if info, err := nodeinfo.Parse(gateName); err == nil {
			if err := nodeinfo.SaveToNode(info, mod.node, true); err != nil {
				mod.log.Error("config error: cannot save nodeinfo %s: %v", gateName, err)
				continue
			}
			gateID = info.Identity
		} else {
			gateID, err = mod.node.Resolver().Resolve(gateName)
			if err != nil {
				mod.log.Error("config error: cannot resolve %s: %v", gateName, err)
				continue
			}
		}

		mod.Subscribe(gateID)
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

	var alwaysLinkedPolicy *policy.AlwaysLinkedPolicy
	if modPolicy, err := modules.Find[*policy.Module](mod.node.Modules()); err == nil {
		alwaysLinkedPolicy = modPolicy.AlwaysLinkedPolicy()
	}

	if alwaysLinkedPolicy != nil {
		alwaysLinkedPolicy.AddIdentity(gateway)
	}

	go func() {
		err := s.Run(mod.ctx)
		if err != nil {
			mod.log.Errorv(1, "gateway %v subscriber ended with error: %v", gateway, err)
		}
		mod.mu.Lock()
		defer mod.mu.Unlock()

		delete(mod.subscribers, hex)
		if alwaysLinkedPolicy != nil {
			alwaysLinkedPolicy.RemoveIdentity(gateway)
		}
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

func (mod *Module) Endpoints() []net.Endpoint {
	var list = make([]net.Endpoint, 0)

	for _, s := range mod.subscribers {
		list = append(list, NewEndpoint(s.Gateway(), mod.node.Identity()))
	}

	return list
}
