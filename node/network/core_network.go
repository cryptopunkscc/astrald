package network

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/debug"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/net"
	"github.com/cryptopunkscc/astrald/node/events"
	"github.com/cryptopunkscc/astrald/node/link"
	"github.com/cryptopunkscc/astrald/tasks"
	"sync"
	"sync/atomic"
	"time"
)

const workers = 16
const queueSize = 64
const logTag = "network"
const defaultQueryTimeout = 30 * time.Second

var _ Network = &CoreNetwork{}
var _ net.Router = &CoreNetwork{}

type CoreNetwork struct {
	links     *LinkSet
	server    *Server
	events    events.Queue
	log       *log.Logger
	node      Node
	tasks     *tasks.FIFOScheduler
	linkTasks map[string]*tasks.Task[net.Link]
	ctx       context.Context
	running   atomic.Bool
	mu        sync.Mutex
	linkMu    sync.Mutex
}

func NewCoreNetwork(node Node, eventParent *events.Queue, log *log.Logger) (*CoreNetwork, error) {
	var err error

	m := &CoreNetwork{
		node:      node,
		log:       log.Tag(logTag),
		links:     NewLinkSet(),
		tasks:     tasks.NewFIFOScheduler(workers, queueSize),
		linkTasks: make(map[string]*tasks.Task[net.Link]),
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
	defer n.running.Store(false)

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

	// run the scheduler
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := n.tasks.Run(ctx); err != nil {
			panic(err)
		}
	}()

	wg.Wait()

	// close all links
	for _, l := range n.links.All() {
		l.Close()
	}

	return nil
}

func (n *CoreNetwork) Server() *Server {
	return n.server
}

func (n *CoreNetwork) Events() *events.Queue {
	return &n.events
}

func (n *CoreNetwork) AddLink(l net.Link) error {
	return n.addLink(l)
}

func (n *CoreNetwork) Links() *LinkSet {
	return n.links
}

func (n *CoreNetwork) addLink(l net.Link) error {
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
