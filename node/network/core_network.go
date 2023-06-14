package network

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/net"
	"github.com/cryptopunkscc/astrald/node/events"
	"github.com/cryptopunkscc/astrald/node/infra"
	"github.com/cryptopunkscc/astrald/node/link"
	"github.com/cryptopunkscc/astrald/node/tracker"
	"github.com/cryptopunkscc/astrald/query"
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
var _ query.Router = &CoreNetwork{}

type CoreNetwork struct {
	links       *LinkSet
	peers       *PeerSet
	server      *Server
	localID     id.Identity
	events      events.Queue
	log         *log.Logger
	tracker     *tracker.CoreTracker
	infra       *infra.CoreInfra
	tasks       *tasks.FIFOScheduler
	linkTasks   map[string]*tasks.Task[*link.Link]
	queryRouter query.Router
	ctx         context.Context
	running     atomic.Bool
	mu          sync.Mutex
}

func NewCoreNetwork(
	localID id.Identity,
	infra *infra.CoreInfra,
	tracker *tracker.CoreTracker,
	eventParent *events.Queue,
	queryRouter query.Router,
	log *log.Logger,
) (*CoreNetwork, error) {
	var err error

	m := &CoreNetwork{
		localID:     localID,
		infra:       infra,
		tracker:     tracker,
		queryRouter: queryRouter,
		log:         log.Tag(logTag),
		peers:       NewPeerSet(),
		links:       NewLinkSet(),
		tasks:       tasks.NewFIFOScheduler(workers, queueSize),
		linkTasks:   make(map[string]*tasks.Task[*link.Link]),
	}

	m.events.SetParent(eventParent)
	m.server, err = newServer(localID, infra, m.AddSecureConn, m.log)
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

	// unlink all peers
	for _, peer := range n.Peers().All() {
		peer.Unlink()
	}

	return nil
}

func (n *CoreNetwork) Server() *Server {
	return n.server
}

func (n *CoreNetwork) Events() *events.Queue {
	return &n.events
}

func (n *CoreNetwork) AddLink(l *link.Link) error {
	return n.addLink(l)
}

// Linkers returns a list of identities we're currently trying to link with
func (n *CoreNetwork) Linkers() []id.Identity {
	n.mu.Lock()
	defer n.mu.Unlock()

	var list = make([]id.Identity, 0, len(n.linkTasks))
	for hex := range n.linkTasks {
		nodeID, _ := id.ParsePublicKeyHex(hex)
		list = append(list, nodeID)
	}

	return list
}

// Peers returns the set of linked peers.
func (n *CoreNetwork) Peers() *PeerSet {
	return n.peers
}

// Link returns a link with the node. If the node is not linked, it will attempt to link to it.
func (n *CoreNetwork) Link(ctx context.Context, nodeID id.Identity) (*link.Link, error) {
	n.mu.Lock()

	// check if peer is already linked
	if peer := n.peers.Find(nodeID); peer != nil {
		if l := peer.PreferredLink(); l != nil {
			n.mu.Unlock()
			return l, nil
		}
	}

	var (
		hexID    = nodeID.PublicKeyHex()
		linkTask *tasks.Task[*link.Link]
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

// AddSecureConn adds a new link to the manager over the provided authenticated connection.
func (n *CoreNetwork) AddSecureConn(conn net.SecureConn) error {
	if !conn.LocalIdentity().IsEqual(n.localID) {
		return ErrIdentityMismatch
	}

	l := link.New(conn, n.log)
	l.SetPriority(NetworkPriority(l.Network()))
	return n.addLink(l)
}

// RequestNewLink schedules a task that will try to establish a new link with the provided node (even if the node
// is already linked).
func (n *CoreNetwork) RequestNewLink(nodeID id.Identity, opts LinkOptions) (*tasks.Task[*link.Link], error) {
	t := tasks.New[*link.Link](&LinkPeerTask{
		RemoteID: nodeID,
		Network:  n,
		options:  opts,
		log:      n.log,
	})

	return t, n.tasks.Add(t)
}

func (n *CoreNetwork) addLink(l *link.Link) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	if !n.running.Load() {
		return ErrNotRunning
	}

	if !l.LocalIdentity().IsEqual(n.localID) {
		return ErrIdentityMismatch
	}

	if err := n.links.Add(l); err != nil {
		return err
	}

	l.Events().SetParent(&n.events)

	var remoteID = l.RemoteIdentity()

	var peer = n.peers.Find(remoteID)
	if peer == nil {
		peer = newPeer(remoteID, &n.events)
		n.peers.Add(peer)
		n.log.Logv(0, "%s linked", l.RemoteIdentity())
		peer.Events().Emit(EventPeerLinked{
			Link: l,
			Peer: peer,
		})
	}

	n.log.Logv(1, "established link with %s over %s", l.RemoteIdentity(), l.Network())
	_ = peer.addLink(l)
	peer.Events().Emit(link.EventLinkEstablished{Link: l})

	// update pings and detect dead links
	peer.Check()

	go func() {
		l.SetQueryRouter(n.queryRouter)
		if err := l.Run(n.ctx); err != nil {
			n.log.Logv(1, "closed link with %s over %s: %s", l.RemoteIdentity(), l.Network(), l.Err())
		}
		if err := n.removeLink(l); err != nil {
			panic(err)
		}
	}()

	return nil
}

func (n *CoreNetwork) removeLink(l *link.Link) error {
	n.mu.Lock()
	defer n.mu.Unlock()

	if err := n.links.Remove(l); err != nil {
		return err
	}

	peer := n.peers.Find(l.RemoteIdentity())
	if peer == nil {
		panic("peer is nil")
	}

	peer.removeLink(l)
	peer.Events().Emit(link.EventLinkClosed{Link: l})

	if peer.links.Count() == 0 {
		n.log.Logv(0, "%s unlinked", l.RemoteIdentity())
		peer.Events().Emit(EventPeerUnlinked{
			Peer: peer,
		})
		n.peers.Remove(peer)
		peer.setUnlinked()
	} else {
		// update pings and detect dead links
		peer.Check()
	}

	return nil
}
