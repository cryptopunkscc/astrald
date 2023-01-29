package peers

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/node/event"
	"github.com/cryptopunkscc/astrald/node/infra"
	"github.com/cryptopunkscc/astrald/node/link"
	"github.com/cryptopunkscc/astrald/node/tracker"
	"github.com/cryptopunkscc/astrald/tasks"
	"io"
	"sync"
)

const workers = 16
const queueSize = 64

type Manager struct {
	Server *Server
	events event.Queue

	pool      *Pool
	localID   id.Identity
	tracker   *tracker.Tracker
	infra     *infra.Infra
	mu        sync.Mutex
	linkQueue chan *link.Link
	tasks     *tasks.FIFOScheduler
	linkTasks map[string]*tasks.Task[*link.Link]
}

func NewManager(
	localID id.Identity,
	infra *infra.Infra,
	tracker *tracker.Tracker,
	eventParent *event.Queue,
) (*Manager, error) {
	var err error

	m := &Manager{
		localID:   localID,
		infra:     infra,
		tracker:   tracker,
		linkQueue: make(chan *link.Link, queueSize),
		tasks:     tasks.NewFIFOScheduler(workers, queueSize),
		linkTasks: make(map[string]*tasks.Task[*link.Link]),
	}

	m.events.SetParent(eventParent)
	m.pool = newPool(localID, &m.events)
	m.Server, err = newServer(localID, infra)
	if err != nil {
		return nil, err
	}

	return m, nil
}

// Run runs the manager until the context is done.
func (m *Manager) Run(ctx context.Context) error {
	linksFromServer, err := m.Server.Run(ctx)
	if err != nil {
		return err
	}

	var wg sync.WaitGroup

	// process queues
	wg.Add(1)
	go func() {
		defer wg.Done()
		for {
			select {
			case <-ctx.Done():
				return

			case l := <-linksFromServer:
				lnk := link.New(l)
				lnk.SetPriority(infra.NetworkPriority(lnk.Network()))
				m.AddLink(lnk)

			case l := <-m.linkQueue:
				go m.runLink(ctx, l)
			}
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
	for peer := range m.All(nil) {
		peer.Unlink()
	}

	return nil
}

func (m *Manager) Queries() <-chan *link.Query {
	return m.pool.Queries()
}

func (m *Manager) Events() *event.Queue {
	return &m.events
}

func (m *Manager) AddLink(l *link.Link) error {
	select {
	case m.linkQueue <- l:
		return nil
	default:
		return errors.New("link queue overflow")
	}
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

// All returns a channel populated with all currently linked peers. If ctx is not nil, the channel will keep
// receiving peers that get linked until ctx is done.
func (m *Manager) All(ctx context.Context) <-chan *Peer {
	return m.pool.Peers(ctx)
}

// Find returns the peer or nil if it's not linked.
func (m *Manager) Find(nodeID id.Identity) *Peer {
	return m.pool.Peer(nodeID)
}

// Link returns a link with the node. If the node is not linked, it will attempt to link to it.
func (m *Manager) Link(ctx context.Context, nodeID id.Identity) (*link.Link, error) {
	m.mu.Lock()

	// check if peer is already linked
	if peer := m.pool.Peer(nodeID); peer != nil {
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

func (m *Manager) runLink(ctx context.Context, l *link.Link) (err error) {
	err = m.pool.addLink(l)
	if err != nil {
		l.Close()
		return
	}

	err = l.Run(ctx)

	m.pool.removeLink(l)

	switch err {
	case nil, // ignore expected errors
		context.Canceled,
		context.DeadlineExceeded,
		io.EOF,
		link.ErrPingTimeout,
		link.ErrIdleTimeout:

	default:
		log.Error("link error: %s", err)
	}

	return
}
