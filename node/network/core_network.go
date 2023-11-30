package network

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/debug"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/net"
	"github.com/cryptopunkscc/astrald/node/events"
	"github.com/cryptopunkscc/astrald/node/link"
	"sync"
	"sync/atomic"
)

const logTag = "network"

var _ Network = &CoreNetwork{}

type CoreNetwork struct {
	ctx     context.Context
	node    Node
	links   *LinkSet
	server  *Server
	events  events.Queue
	log     *log.Logger
	running atomic.Bool
	mu      sync.Mutex
}

func NewCoreNetwork(node Node, eventParent *events.Queue, log *log.Logger) (*CoreNetwork, error) {
	var err error

	m := &CoreNetwork{
		node:  node,
		log:   log.Tag(logTag),
		links: NewLinkSet(),
	}

	m.events.SetParent(eventParent)
	m.server, err = newServer(node.Identity(), node.Infra(), m.AddLink, m.log)
	if err != nil {
		return nil, err
	}

	return m, nil
}

// Run runs the manager until the context is done.
func (n *CoreNetwork) Run(ctx context.Context) error {
	if !n.running.CompareAndSwap(false, true) {
		return errors.New("already running")
	}

	n.ctx = ctx
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer debug.SaveLog(debug.SigInt)
		defer wg.Done()

		err := n.server.Run(ctx)
		switch {
		case err == nil:
		case errors.Is(err, context.Canceled):
		default:
			n.log.Error("server error: %s", err)
		}

	}()

	<-ctx.Done()

	// wait for the server to shut down
	wg.Wait()

	n.mu.Lock()
	defer n.mu.Unlock()

	// close all links
	for _, l := range n.links.All() {
		l.Close()
	}

	n.running.Store(false)

	return nil
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

	if corelink, ok := l.(*link.CoreLink); ok {
		corelink.SetUplink(n.node.Router())
		defer corelink.Check()
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

	n.log.Logv(1, "added link %v with %v", active.ID(), l.RemoteIdentity())
	n.events.Emit(EventLinkAdded{Link: active})

	return nil
}

func (n *CoreNetwork) Server() *Server {
	return n.server
}

func (n *CoreNetwork) Events() *events.Queue {
	return &n.events
}

func (n *CoreNetwork) Links() *LinkSet {
	return n.links
}
