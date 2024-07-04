package policy

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/mod/nodes"
	"github.com/cryptopunkscc/astrald/net"
	"github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/node/network"
	"github.com/cryptopunkscc/astrald/node/tracker"
	"sync"
	"time"
)

var _ Policy = &OptimizeLinksPolicy{}

// OptimizeLinksPolicy makes sure we're linked via the best possible network
type OptimizeLinksPolicy struct {
	*Module
	ctx     context.Context
	workers map[string]*optimizeLinksWorker
	mu      sync.Mutex
}

func NewOptimizeLinksPolicy(mod *Module) *OptimizeLinksPolicy {
	return &OptimizeLinksPolicy{
		Module:  mod,
		workers: make(map[string]*optimizeLinksWorker),
	}
}

func (policy *OptimizeLinksPolicy) Run(ctx context.Context) error {
	policy.ctx = ctx

	events := policy.node.Events().Subscribe(ctx)

	for event := range events {
		event, ok := event.(network.EventLinkAdded)
		if !ok {
			continue
		}

		policy.addWorker(event.Link.RemoteIdentity())
	}
	return nil
}

func (policy *OptimizeLinksPolicy) addWorker(target id.Identity) error {
	policy.mu.Lock()
	defer policy.mu.Unlock()

	if target.IsZero() {
		return errors.New("target cannot be zero")
	}

	var hex = target.PublicKeyHex()
	var worker = newOptimizeLinksWorker(policy.Module, target)

	policy.workers[hex] = worker

	go func() {
		worker.Run(policy.ctx)

		policy.mu.Lock()
		defer policy.mu.Unlock()

		delete(policy.workers, hex)
	}()

	return nil
}

func (policy *OptimizeLinksPolicy) Name() string {
	return "optimize_links"
}

type optimizeLinksWorker struct {
	*Module
	node   node.Node
	target id.Identity
	cond   *sync.Cond
}

func newOptimizeLinksWorker(mod *Module, target id.Identity) *optimizeLinksWorker {
	return &optimizeLinksWorker{
		Module: mod,
		node:   mod.node,
		target: target,
		cond:   sync.NewCond(&sync.Mutex{}),
	}
}

func (worker *optimizeLinksWorker) Run(ctx context.Context) error {
	worker.cond.L.Lock()
	defer worker.cond.L.Unlock()

	go func() {
		<-ctx.Done()
		worker.cond.Broadcast()
	}()

	go worker.watch(ctx)

	for {
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		var links = worker.node.Network().Links().ByRemoteIdentity(worker.target).All()

		if len(links) == 0 {
			return nil
		}

		endpoints, err := worker.node.Tracker().EndpointsByIdentity(worker.target)
		if err != nil {
			return err
		}

		var _, bestScore = bestLinkScore(links)
		var try = make([]net.Endpoint, 0)
		for _, ep := range endpoints {
			if scoreNetwork(ep.Network()) > bestScore {
				try = append(try, ep)
			}
		}

		if len(try) == 0 {
			worker.cond.Wait()
			continue
		}

		_, err = worker.nodes.Link(ctx, worker.target, nodes.LinkOpts{Endpoints: try})
		if err == nil {
			time.Sleep(1 * time.Second)
			continue
		}

		worker.cond.Wait()
	}
}

func (worker *optimizeLinksWorker) watch(ctx context.Context) {
	events := worker.node.Events().Subscribe(ctx)

	for {
		select {
		case event, ok := <-events:
			if !ok {
				return
			}

			switch event := event.(type) {
			case network.EventLinkAdded:
				if event.Link.RemoteIdentity().IsEqual(worker.target) {
					worker.cond.Broadcast()
				}

			case network.EventLinkRemoved:
				if event.Link.RemoteIdentity().IsEqual(worker.target) {
					time.Sleep(250 * time.Millisecond) // wait in case all links are being closed
					worker.cond.Broadcast()
				}

			case tracker.EventNewEndpoint:
				if event.Identity.IsEqual(worker.target) {
					worker.cond.Broadcast()
				}
			}

		case <-ctx.Done():
			return
		}
	}
}

func scoreNetwork(network string) int {
	switch network {
	case "tor":
		return 10
	case "bt":
		return 20
	case "gw":
		return 30
	case "inet", "tcp":
		return 40
	}
	return 0
}

func bestLinkScore(links []*network.ActiveLink) (*network.ActiveLink, int) {
	if len(links) == 0 {
		return nil, 0
	}

	var best = links[0]
	var bestScore = scoreNetwork(linkNetwork(best))

	for _, lnk := range links {
		s := scoreNetwork(linkNetwork(lnk.Link))
		if s > bestScore {
			best = lnk
			bestScore = s
		}
	}
	return best, bestScore
}

func linkNetwork(l net.Link) string {
	var t = l.Transport()
	if t == nil {
		return ""
	}
	if t.LocalEndpoint() != nil {
		return t.LocalEndpoint().Network()
	}
	if t.RemoteEndpoint() != nil {
		return t.RemoteEndpoint().Network()
	}
	return ""
}
