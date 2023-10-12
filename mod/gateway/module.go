package gateway

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/net"
	"github.com/cryptopunkscc/astrald/node"
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

	for _, gate := range mod.config.Subscribe {
		gateID, err := mod.node.Resolver().Resolve(gate)
		if err != nil {
			mod.log.Error("config error: cannot resolve %s: %v", gate, err)
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

	var hex = gateway.PublicKeyHex()

	if _, found := mod.subscribers[hex]; found {
		return errors.New("already subscribed")
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
		return errors.New("subscription not found")
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
