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
	"io"
	"log"
	"sync"
)

const concurrency = 16

type Manager struct {
	Pool   *Pool
	Server *Server

	localID   id.Identity
	tracker   *tracker.Tracker
	infra     *infra.Infra
	linkers   map[string]*Linker
	mu        sync.Mutex
	events    event.Queue
	links     chan *alink.Link
	linkQueue chan *link.Link
}

func NewManager(localID id.Identity, infra *infra.Infra, tracker *tracker.Tracker, eventParent *event.Queue) (*Manager, error) {
	var err error

	m := &Manager{
		localID:   localID,
		linkers:   make(map[string]*Linker),
		infra:     infra,
		tracker:   tracker,
		links:     make(chan *alink.Link, 16),
		linkQueue: make(chan *link.Link, 16),
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
	linksFromServer, err := m.Server.Run(ctx)
	if err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return nil

		case l := <-linksFromServer:
			m.AddLink(link.New(l))

		case l := <-m.linkQueue:
			go m.runLink(ctx, l)
		}
	}
}

func (m *Manager) Queries() <-chan *link.Query {
	return m.Pool.Queries()
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

func (m *Manager) Link(ctx context.Context, remoteID id.Identity) (*link.Link, error) {
	linker, isNew := m.getLinker(remoteID)

	if !isNew {
		select {
		case <-linker.Done():
			if linker.Error() != nil {
				return nil, linker.Error()
			}
			return linker.link, nil

		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}

	defer close(linker.done)

	ctx, done := context.WithCancel(ctx)

	type res struct {
		link *link.Link
		err  error
	}

	var ch = make(chan res, 2)
	var wg sync.WaitGroup
	wg.Add(2)

	// observe external link sources
	go func() {
		defer wg.Done()
		for e := range m.events.Subscribe(ctx) {
			e, ok := e.(link.EventLinkEstablished)
			if !ok {
				continue
			}
			if !e.Link.RemoteIdentity().IsEqual(remoteID) {
				continue
			}

			ch <- res{e.Link, nil}
			return
		}
	}()

	// try to make a new link
	go func() {
		defer wg.Done()
		l, err := m.getOrMakeLink(ctx, remoteID)
		ch <- res{l, err}
	}()

	// wait for the result
	r := <-ch
	linker.link, linker.err = r.link, r.err

	// clean up
	done()
	wg.Wait()
	close(ch)
	m.deleteLinker(remoteID)

	return linker.link, linker.err
}

func (m *Manager) runLink(ctx context.Context, l *link.Link) (err error) {
	err = m.Pool.addLink(l)
	if err != nil {
		l.Close()
		return
	}

	err = l.Run(ctx)

	m.Pool.removeLink(l)

	switch err {
	case nil, // ignore expected errors
		context.Canceled,
		context.DeadlineExceeded,
		io.EOF,
		link.ErrPingTimeout,
		link.ErrIdleTimeout:

	default:
		log.Println("link error:", err)
	}

	return
}

func (m *Manager) getLinker(remoteID id.Identity) (linker *Linker, new bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if linker, found := m.linkers[remoteID.String()]; found {
		return linker, false
	}

	linker = &Linker{
		remoteID: remoteID,
		done:     make(chan struct{}),
	}

	m.linkers[remoteID.String()] = linker

	return linker, true
}

func (m *Manager) deleteLinker(remoteID id.Identity) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	if _, found := m.linkers[remoteID.String()]; !found {
		return errors.New("linker not found")
	}

	delete(m.linkers, remoteID.String())
	return nil
}

func (m *Manager) getOrMakeLink(ctx context.Context, remoteID id.Identity) (*link.Link, error) {
	// check if peer is already connected
	if peer := m.Pool.Peer(remoteID); peer != nil {
		if l := peer.PreferredLink(); l != nil {
			return l, nil
		}
	}

	rawLink, err := m.makeNewLink(ctx, remoteID)
	if err != nil {
		return nil, err
	}

	l := link.New(rawLink)

	// add the link to the pool
	if err := m.AddLink(l); err != nil {
		return nil, err
	}

	return l, nil
}

func (m *Manager) makeNewLink(ctx context.Context, remoteID id.Identity) (*alink.Link, error) {
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
