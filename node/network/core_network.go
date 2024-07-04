package network

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/debug"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/net"
	"github.com/cryptopunkscc/astrald/node/events"
	"github.com/cryptopunkscc/astrald/sig"
	"sync"
	"sync/atomic"
)

const logTag = "network"

var _ Network = &CoreNetwork{}

type CoreNetwork struct {
	ctx     context.Context
	node    Node
	linkers sig.Set[Linker]
	links   *LinkSet
	events  events.Queue
	log     *log.Logger
	running atomic.Bool
	mu      sync.Mutex
}

func NewCoreNetwork(node Node, eventParent *events.Queue, log *log.Logger) (*CoreNetwork, error) {
	m := &CoreNetwork{
		node:  node,
		log:   log.Tag(logTag),
		links: NewLinkSet(),
	}

	m.events.SetParent(eventParent)

	return m, nil
}

// Run runs the manager until the context is done.
func (n *CoreNetwork) Run(ctx context.Context) error {
	if !n.running.CompareAndSwap(false, true) {
		return errors.New("already running")
	}
	n.ctx = ctx

	<-ctx.Done()

	n.mu.Lock()
	defer n.mu.Unlock()

	n.log.Logv(1, "closing all links...")

	var wg sync.WaitGroup
	// close all links
	for _, l := range n.links.All() {
		wg.Add(1)
		go func() {
			<-l.Done()
			wg.Done()
		}()
		l.Close()
	}
	wg.Wait()

	n.running.Store(false)

	return nil
}

func (n *CoreNetwork) AddLinker(linker Linker) error {
	return n.linkers.Add(linker)
}

func (n *CoreNetwork) Link(ctx context.Context, identity id.Identity) (net.Link, error) {
	linkers := n.linkers.Clone()

	if len(linkers) == 0 {
		return nil, errors.New("no linkers found")
	}

	var errs = make(chan error, len(linkers))
	var link = make(chan net.Link, 1)

	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	var wg sync.WaitGroup

	wg.Add(len(linkers))
	for _, linker := range linkers {
		linker := linker
		go func() {
			defer wg.Done()

			l, err := linker.Link(ctx, identity)
			if err != nil {
				errs <- err
				return
			}

			select {
			case link <- l:
				cancel()
			default:
				l.Close()
			}
		}()
	}

	wg.Wait()
	close(errs)

	if len(link) != 1 {
		return nil, errors.Join(sig.ChanToArray(errs)...)
	}

	l := <-link

	err := n.AddLink(l)
	if err != nil {
		l.Close()
		return nil, err
	}

	return l, nil
}

func (n *CoreNetwork) AddLink(l net.Link) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	if !n.running.Load() {
		return ErrNotRunning
	}

	if !l.LocalIdentity().IsEqual(n.node.Identity()) {
		return ErrIdentityMismatch
	}

	active, err := n.links.Add(l)
	if err != nil {
		return err
	}

	n.node.Router().AddRoute(l.LocalIdentity(), l.RemoteIdentity(), l, 50)

	go func() {
		defer debug.SaveLog(debug.SigInt)

		err := l.Run(n.ctx)
		n.node.Router().RemoveRoute(l.LocalIdentity(), l.RemoteIdentity(), l)
		if e := n.links.Remove(active.ID()); e != nil {
			panic(e)
		}
		n.log.Logv(2, "removed link %v with %v: %v", active.ID(), l.RemoteIdentity(), err)
		n.events.Emit(EventLinkRemoved{Link: active})
	}()

	n.log.Logv(1, "added link %v with %v (%s)", active.ID(), l.RemoteIdentity(), net.Network(l))
	n.events.Emit(EventLinkAdded{Link: active})

	return nil
}

func (n *CoreNetwork) Events() *events.Queue {
	return &n.events
}

func (n *CoreNetwork) Links() *LinkSet {
	return n.links
}
