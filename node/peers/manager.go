package peers

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/auth/id"
	alink "github.com/cryptopunkscc/astrald/link"
	"github.com/cryptopunkscc/astrald/node/contacts"
	"github.com/cryptopunkscc/astrald/node/event"
	"github.com/cryptopunkscc/astrald/node/infra"
	"github.com/cryptopunkscc/astrald/node/link"
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

	localID  id.Identity
	contacts *contacts.Manager
	infra    *infra.Infra
	linkers  map[string]*Linker
	mu       sync.Mutex
	events   event.Queue
}

func Run(ctx context.Context, localID id.Identity, infra *infra.Infra, contacts *contacts.Manager, eventParent *event.Queue) (*Manager, error) {
	var err error

	m := &Manager{
		localID:  localID,
		linkers:  make(map[string]*Linker),
		infra:    infra,
		contacts: contacts,
	}

	m.events.SetParent(eventParent)
	m.Pool = NewPool(localID, &m.events)

	m.Server, err = runServer(ctx, localID, infra)
	if err != nil {
		return nil, err
	}

	go func() {
		for l := range m.Server.Links() {
			if err := m.AddLink(l); err != nil {
				panic(err)
			}
		}
		log.Println("peers: server done")
	}()

	return m, nil
}

func (m *Manager) Queries() <-chan *link.Query {
	return m.Pool.Queries()
}

func (m *Manager) AddLink(l *alink.Link) (err error) {
	lnk := link.Wrap(l, &m.events)
	err = m.Pool.AddLink(lnk)

	// set up a timeout for the link
	if err == nil {
		linkCtx, cancel := context.WithCancel(context.Background())
		sig.On(lnk, cancel)
		sig.OnCtx(linkCtx, sig.Idle(linkCtx, lnk, LinkTimeout), func() {
			log.Println("peers.Manager.AddLink(): closing link due to timeout")
			lnk.Close()
		})
	}
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

	contact := m.contacts.Find(remoteID, false)
	if contact == nil {
		return nil, errors.New("no address found")
	}

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
			contact.Addr(),
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
