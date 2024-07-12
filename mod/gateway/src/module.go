package gateway

import (
	"context"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/core"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/mod/dir"
	"github.com/cryptopunkscc/astrald/mod/discovery"
	"github.com/cryptopunkscc/astrald/mod/nodes"
	"github.com/cryptopunkscc/astrald/mod/policy"
	"github.com/cryptopunkscc/astrald/net"
	node2 "github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/tasks"
	"sync"
)

const NetworkName = "gw"

type Module struct {
	config      Config
	node        node2.Node
	log         *log.Logger
	ctx         context.Context
	dialer      *Dialer
	subscribers map[string]*Subscriber
	mu          sync.Mutex
	sdp         discovery.Module
	policy      policy.Module
	nodes       nodes.Module
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

	mod.sdp, _ = core.Load[discovery.Module](mod.node, discovery.ModuleName)
	mod.policy, _ = core.Load[policy.Module](mod.node, policy.ModuleName)

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
			gateID, err = mod.node.Resolver().Resolve(gateName)
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

	if mod.policy != nil {
		mod.policy.AddAlwaysLinkedIdentity(gateway)
	}

	go func() {
		err := s.Run(mod.ctx)
		if err != nil {
			mod.log.Errorv(1, "gateway %v subscriber ended with error: %v", gateway, err)
		}
		mod.mu.Lock()
		defer mod.mu.Unlock()

		delete(mod.subscribers, hex)
		if mod.policy != nil {
			mod.policy.RemoveAlwaysLinkedIdentity(gateway)
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
