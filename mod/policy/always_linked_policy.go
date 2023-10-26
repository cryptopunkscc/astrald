package policy

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/node/link"
	"sync"
	"sync/atomic"
	"time"
)

var _ Policy = &AlwaysLinkedPolicy{}

// AlwaysLinkedPolicy keeps the node linked to specified targets whenever possible.
type AlwaysLinkedPolicy struct {
	*Module
	ctx     context.Context
	workers map[string]*alwaysLinkedWorker
	mu      sync.Mutex
}

func NewAlwaysLinkedPolicy(module *Module) *AlwaysLinkedPolicy {
	policy := &AlwaysLinkedPolicy{
		Module:  module,
		workers: make(map[string]*alwaysLinkedWorker),
	}
	policy.mu.Lock()
	return policy
}

func (policy *AlwaysLinkedPolicy) Run(ctx context.Context) error {
	policy.ctx = ctx
	policy.mu.Unlock()

	<-ctx.Done()
	return nil
}

func (policy *AlwaysLinkedPolicy) AddIdentity(identity id.Identity) error {
	policy.mu.Lock()
	defer policy.mu.Unlock()

	if policy.ctx == nil {
		return errors.New("policy not running")
	}

	if identity.IsZero() {
		return errors.New("identity cannot be zero")
	}

	hex := identity.PublicKeyHex()

	if w, found := policy.workers[hex]; found {
		w.Add(1)
		return nil
	}

	worker := newAlwaysLinkedWorker(policy.node, identity)
	policy.workers[hex] = worker

	go func() {
		if err := worker.Run(policy.ctx); err != nil {
			policy.log.Errorv(2, "always_linked_policy worker ended with error:", err)
		}
		policy.mu.Lock()
		defer policy.mu.Unlock()
		delete(policy.workers, hex)
	}()

	return nil
}

func (policy *AlwaysLinkedPolicy) RemoveIdentity(identity id.Identity) error {
	policy.mu.Lock()
	defer policy.mu.Unlock()

	if identity.IsZero() {
		return errors.New("identity cannot be zero")
	}

	hex := identity.PublicKeyHex()

	worker, found := policy.workers[hex]
	if !found {
		return errors.New("identity not found")
	}

	worker.Add(-1)
	return nil
}

func (policy *AlwaysLinkedPolicy) Identities() []id.Identity {
	policy.mu.Lock()
	defer policy.mu.Unlock()

	var list = make([]id.Identity, 0)
	for _, worker := range policy.workers {
		list = append(list, worker.target)
	}

	return list
}

func (policy *AlwaysLinkedPolicy) Name() string {
	return "always_linked"
}

type alwaysLinkedWorker struct {
	node     node.Node
	target   id.Identity
	errCount int
	cancel   context.CancelFunc
	counter  atomic.Int32
}

func newAlwaysLinkedWorker(node node.Node, target id.Identity) *alwaysLinkedWorker {
	w := &alwaysLinkedWorker{
		node:   node,
		target: target,
	}
	w.counter.Add(1)
	return w
}

func (worker *alwaysLinkedWorker) Add(delta int32) {
	if worker.counter.Add(delta) == 0 {
		worker.cancel()
	}
}

func (worker *alwaysLinkedWorker) Run(ctx context.Context) error {
	ctx, worker.cancel = context.WithCancel(ctx)

	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		var links = worker.node.Network().Links().ByRemoteIdentity(worker.target).All()

		if len(links) > 0 {
			worker.errCount = 0
			var wg sync.WaitGroup
			wg.Add(len(links))
			for _, lnk := range links {
				lnk := lnk
				go func() {
					defer wg.Done()
					select {
					case <-lnk.Done():
					case <-ctx.Done():
					}
				}()
			}
			wg.Wait()
			continue
		}

		lnk, err := link.MakeLink(ctx, worker.node, worker.target, link.Opts{})
		if err != nil {
			worker.errCount++

			select {
			case <-time.After(retryIvals.At(worker.errCount - 1)):
				continue

			case <-ctx.Done():
				return ctx.Err()
			}
		}

		worker.node.Network().AddLink(lnk)
	}
}
