package peers

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/auth/id"
	ainfra "github.com/cryptopunkscc/astrald/infra"
	alink "github.com/cryptopunkscc/astrald/link"
	"github.com/cryptopunkscc/astrald/node/event"
	"github.com/cryptopunkscc/astrald/node/infra"
	"github.com/cryptopunkscc/astrald/node/link"
	"github.com/cryptopunkscc/astrald/node/tracker"
	"github.com/cryptopunkscc/astrald/sig"
	"log"
	"sync"
	"time"
)

const LinkTimeout = 5 * time.Minute
const concurrency = 16

type Manager struct {
	Pool   *Pool
	Server *Server

	localID id.Identity
	tracker *tracker.Tracker
	infra   *infra.Infra
	linkers map[string]*Linker
	mu      sync.Mutex
	events  event.Queue
	links   chan *alink.Link
	timeout chan *link.Link
}

func NewManager(localID id.Identity, infra *infra.Infra, tracker *tracker.Tracker, eventParent *event.Queue) (*Manager, error) {
	var err error

	m := &Manager{
		localID: localID,
		linkers: make(map[string]*Linker),
		infra:   infra,
		tracker: tracker,
		links:   make(chan *alink.Link, 16),
		timeout: make(chan *link.Link, 16),
	}

	m.events.SetParent(eventParent)
	m.Pool = NewPool(localID, &m.events)
	m.Server, err = newServer(localID, infra)
	if err != nil {
		return nil, err
	}

	return m, nil
}

func (m *Manager) Run(ctx context.Context) error {
	links, err := m.Server.Run(ctx)
	if err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return nil

		case link := <-links:
			m.AddLink(link)

		case link := <-m.timeout:
			// set up a timeout for the link
			linkCtx, cancel := context.WithCancel(ctx)
			sig.On(link, cancel)
			sig.OnCtx(linkCtx, sig.Idle(linkCtx, link, LinkTimeout), func() {
				log.Println("peers.Manager.AddLink(): closing link due to timeout")
				link.Close()
			})
		}
	}
}

func (m *Manager) Queries() <-chan *link.Query {
	return m.Pool.Queries()
}

func (m *Manager) AddLink(l *alink.Link) (err error) {
	nodeLink := link.Wrap(l, &m.events)
	err = m.Pool.AddLink(nodeLink)

	if err != nil {
		nodeLink.Close()
	}

	m.timeout <- nodeLink

	return
}

func (m *Manager) Linkers() <-chan *Linker {
	m.mu.Lock()
	defer m.mu.Unlock()

	ch := make(chan *Linker, len(m.linkers))
	for _, l := range m.linkers {
		ch <- l
	}
	close(ch)

	return ch
}

func (m *Manager) Link(ctx context.Context, remoteID id.Identity) (*Peer, error) {
	m.mu.Lock()

	// try attaching to the existing linker first
	if linker, found := m.linkers[remoteID.String()]; found {
		m.mu.Unlock()

		select {
		case <-linker.Done():
			if linker.Error() != nil {
				return nil, linker.Error()
			}
			if peer := m.Pool.Peer(remoteID); peer != nil {
				return peer, nil
			}
			return nil, errors.New("unexpected error")

		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	// check if peer is already connected
	if peer := m.Pool.Peer(remoteID); peer != nil {
		m.mu.Unlock()
		return peer, nil
	}

	// create a context for the linker
	lctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// start a new linker
	linker := &Linker{ctx: lctx, remoteID: remoteID}
	m.linkers[remoteID.String()] = linker
	m.mu.Unlock()

	defer func() {
		m.mu.Lock()
		delete(m.linkers, remoteID.String())
		m.mu.Unlock()
	}()

	link, err := m.makeLink(ctx, remoteID)
	if err != nil {
		linker.err = err
		return nil, err
	}

	// add the link to the pool
	if err := m.AddLink(link); err != nil {
		linker.err = err
		return nil, err
	}

	if peer := m.Pool.Peer(remoteID); peer != nil {
		return peer, nil
	}

	linker.err = errors.New("unexpected error")
	return nil, linker.err
}

func (m *Manager) makeLink(ctx context.Context, remoteID id.Identity) (*alink.Link, error) {
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	// Fetch addresses for the remote identity
	addrs, err := m.tracker.AddrByIdentity(remoteID)
	if err != nil {
		return nil, err
	}
	if len(addrs) == 0 {
		return nil, errors.New("identity has no addresses")
	}

	// Populate a channel with addresses
	addrCh := make(chan ainfra.Addr, len(addrs))
	for _, a := range addrs {
		addrCh <- a
	}
	close(addrCh)

	authed := NewConcurrentHandshake(
		m.localID,
		remoteID,
		concurrency,
	).Outbound(
		ctx,
		NewConcurrentDialer(
			m.infra,
			concurrency,
		).Dial(
			ctx,
			addrCh,
		),
	)

	defer func() {
		go func() {
			for a := range authed {
				a.Close()
			}
		}()
	}()

	if a, ok := <-authed; ok {
		return alink.New(a), nil
	}

	return nil, errors.New("linking failed")
}
