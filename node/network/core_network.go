package network

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/auth/id"
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
	m.server, err = newServer(node.Identity(), node.Infra(), func(conn net.SecureConn) error {
		return m.AddLink(link.NewCoreLink(conn))
	}, m.log)
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

// Link returns a link with the node. If the node is not linked, it will attempt to link to it.
func (n *CoreNetwork) Link(ctx context.Context, nodeID id.Identity) (net.Link, error) {
	n.mu.Lock()

	// check if peer is already linked
	var links = n.links.ByRemoteIdentity(nodeID).All()
	if len(links) > 0 {
		n.mu.Unlock()
		return links[0], nil
	}

	var (
		hexID    = nodeID.PublicKeyHex()
		linkTask *tasks.Task[net.Link]
		ok       bool
	)

	// use the link task that's already running for this node, or start one
	linkTask, ok = n.linkTasks[hexID]
	if !ok {
		newTask, err := n.RequestNewLink(nodeID, LinkOptions{})
		if err != nil {
			n.mu.Unlock()
			return nil, err
		}

		linkTask = newTask
		n.linkTasks[hexID] = newTask

		go func() {
			<-linkTask.Done()
			n.mu.Lock()
			delete(n.linkTasks, hexID)
			n.mu.Unlock()
		}()
	}
	n.mu.Unlock()

	// wait for the task to finish, or the context to end
	select {
	case <-ctx.Done():
		return nil, ctx.Err()

	case <-linkTask.Done():
		return linkTask.Result(), linkTask.Err()
	}
}

// RequestNewLink schedules a task that will try to establish a new link with the provided node (even if the node
// is already linked).
func (n *CoreNetwork) RequestNewLink(nodeID id.Identity, opts LinkOptions) (*tasks.Task[net.Link], error) {
	t := tasks.New[net.Link](&LinkPeerTask{
		RemoteID: nodeID,
		Network:  n,
		options:  opts,
		log:      n.log,
	})

	return t, n.tasks.Add(t)
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

	if err := n.links.Add(l); err != nil {
		return err
	}

	if corelink, ok := l.(*link.CoreLink); ok {
		corelink.SetUplink(n.node.Router())
		defer corelink.Check()
	}

	n.log.Logv(1, "added link with %v", l.RemoteIdentity())
	n.events.Emit(EventLinkAdded{Link: l})

	go func() {
		err := l.Run(n.ctx)
		if e := n.links.Remove(l); e != nil {
			panic(e)
		}
		n.log.Logv(2, "removed link with %v: %v", l.RemoteIdentity(), err)
		n.events.Emit(EventLinkRemoved{Link: l})
	}()

	return nil
}
