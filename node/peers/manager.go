package peers

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/auth"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/node/event"
	"github.com/cryptopunkscc/astrald/node/infra"
	"github.com/cryptopunkscc/astrald/node/link"
	"github.com/cryptopunkscc/astrald/node/tracker"
	"github.com/cryptopunkscc/astrald/tasks"
	"sync"
	"sync/atomic"
)

const workers = 16
const queueSize = 64

type Manager struct {
	links        *LinkSet
	peers        *PeerSet
	server       *Server
	localID      id.Identity
	events       event.Queue
	tracker      *tracker.Tracker
	infra        *infra.Infra
	tasks        *tasks.FIFOScheduler
	linkTasks    map[string]*tasks.Task[*link.Link]
	queryHandler link.QueryHandlerFunc
	ctx          context.Context
	running      atomic.Bool
	mu           sync.Mutex
}

func NewManager(
	localID id.Identity,
	infra *infra.Infra,
	tracker *tracker.Tracker,
	eventParent *event.Queue,
	queryHandler link.QueryHandlerFunc,
) (*Manager, error) {
	var err error

	m := &Manager{
		localID:      localID,
		infra:        infra,
		tracker:      tracker,
		queryHandler: queryHandler,
		peers:        NewPeerSet(),
		links:        NewLinkSet(),
		tasks:        tasks.NewFIFOScheduler(workers, queueSize),
		linkTasks:    make(map[string]*tasks.Task[*link.Link]),
	}

	m.events.SetParent(eventParent)
	m.server, err = newServer(localID, infra, m.AddAuthConn)
	if err != nil {
		return nil, err
	}

	return m, nil
}

// Run runs the manager until the context is done.
func (m *Manager) Run(ctx context.Context) error {
	if !m.running.CompareAndSwap(false, true) {
		return errors.New("already running")
	}
	defer m.running.Store(false)

	m.ctx = ctx
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()

		err := m.server.Run(ctx)
		if err != nil {
			log.Error("server error: %s", err)
		}
	}()

	// run the scheduler
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := m.tasks.Run(ctx); err != nil {
			panic(err)
		}
	}()

	wg.Wait()

	// unlink all peers
	for _, peer := range m.Peers() {
		peer.Unlink()
	}

	return nil
}

func (m *Manager) Server() *Server {
	return m.server
}

func (m *Manager) Events() *event.Queue {
	return &m.events
}

func (m *Manager) AddLink(l *link.Link) error {
	return m.addLink(l)
}

// Linkers returns a list of identities we're currently trying to link with
func (m *Manager) Linkers() <-chan id.Identity {
	m.mu.Lock()
	defer m.mu.Unlock()

	ch := make(chan id.Identity, len(m.linkTasks))
	for hex := range m.linkTasks {
		nodeID, _ := id.ParsePublicKeyHex(hex)
		ch <- nodeID
	}
	close(ch)

	return ch
}

// Peers returns a list of all linked peers.
func (m *Manager) Peers() []*Peer {
	return m.peers.All()
}

// Find returns the peer or nil if it's not linked.
func (m *Manager) Find(nodeID id.Identity) *Peer {
	return m.peers.Find(nodeID)
}

// Link returns a link with the node. If the node is not linked, it will attempt to link to it.
func (m *Manager) Link(ctx context.Context, nodeID id.Identity) (*link.Link, error) {
	m.mu.Lock()

	// check if peer is already linked
	if peer := m.peers.Find(nodeID); peer != nil {
		if l := peer.PreferredLink(); l != nil {
			m.mu.Unlock()
			return l, nil
		}
	}

	var (
		hexID    = nodeID.PublicKeyHex()
		linkTask *tasks.Task[*link.Link]
		ok       bool
	)

	// use the link task that's already running for this node, or start one
	linkTask, ok = m.linkTasks[hexID]
	if !ok {
		newTask, err := m.RequestNewLink(nodeID, LinkOptions{})
		if err != nil {
			m.mu.Unlock()
			return nil, err
		}

		linkTask = newTask
		m.linkTasks[hexID] = newTask

		go func() {
			<-linkTask.Done()
			m.mu.Lock()
			delete(m.linkTasks, hexID)
			m.mu.Unlock()
		}()
	}
	m.mu.Unlock()

	// wait for the task to finish, or the context to end
	select {
	case <-ctx.Done():
		return nil, ctx.Err()

	case <-linkTask.Done():
		return linkTask.Result(), linkTask.Err()
	}
}

// AddAuthConn adds a new link to the manager over the provided authenticated connection.
func (m *Manager) AddAuthConn(conn auth.Conn) error {
	if !conn.LocalIdentity().IsEqual(m.localID) {
		return ErrIdentityMismatch
	}

	l := link.New(conn)
	l.SetPriority(infra.NetworkPriority(l.Network()))
	return m.addLink(l)
}

// RequestNewLink schedules a task that will try to establish a new link with the provided node (even if the node
// is already linked).
func (m *Manager) RequestNewLink(nodeID id.Identity, opts LinkOptions) (*tasks.Task[*link.Link], error) {
	t := tasks.New[*link.Link](&LinkPeerTask{
		RemoteID: nodeID,
		Peers:    m,
		options:  opts,
	})

	return t, m.tasks.Add(t)
}

func (m *Manager) onQuery(query *link.Query) (err error) {
	err = m.queryHandler(query)

	return err
}

func (m *Manager) addLink(l *link.Link) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if !m.running.Load() {
		return ErrNotRunning
	}

	if !l.LocalIdentity().IsEqual(m.localID) {
		return ErrIdentityMismatch
	}

	if err := m.links.Add(l); err != nil {
		return err
	}

	l.Events().SetParent(&m.events)

	var remoteID = l.RemoteIdentity()

	var peer = m.peers.Find(remoteID)
	if peer == nil {
		peer = newPeer(remoteID, &m.events)
		m.peers.Add(peer)
		peer.Events().Emit(EventPeerLinked{
			Link: l,
			Peer: peer,
		})
	}

	_ = peer.addLink(l)
	peer.Events().Emit(link.EventLinkEstablished{Link: l})

	go func() {
		l.SetQueryHandler(m.onQuery)
		if err := l.Run(m.ctx); err != nil {
			log.Error("link closed: %s", err)
		}
		if err := m.removeLink(l); err != nil {
			panic(err)
		}
	}()

	return nil
}

func (m *Manager) removeLink(l *link.Link) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if err := m.links.Remove(l); err != nil {
		return err
	}

	peer := m.peers.Find(l.RemoteIdentity())
	if peer == nil {
		panic("peer is nil")
	}

	peer.removeLink(l)
	peer.Events().Emit(link.EventLinkClosed{Link: l})

	if peer.links.Count() == 0 {
		peer.Events().Emit(EventPeerUnlinked{
			Peer: peer,
		})
		m.peers.Remove(peer)
		peer.setUnlinked()
	}

	return nil
}
