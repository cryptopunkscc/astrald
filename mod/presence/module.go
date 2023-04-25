package presence

import (
	"context"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/infra"
	"github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/node/event"
	"github.com/cryptopunkscc/astrald/sig"
	"sync"
	"time"
)

type Module struct {
	node   node.Node
	config Config

	entries map[string]*entry
	mu      sync.Mutex
	events  event.Queue
	skip    map[string]struct{}
}

func (m *Module) Run(ctx context.Context) (err error) {
	ctx, shutdown := context.WithCancel(ctx)

	var errCh = make(chan error, 2)
	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()

		discover, err := m.node.Infra().Discover(ctx)
		if err != nil {
			errCh <- err
			return
		}

		for presence := range discover {
			hex := presence.Identity.PublicKeyHex()
			if _, skip := m.skip[hex]; skip {
				continue
			}

			m.handle(ctx, presence)
		}
	}()

	wg.Add(1)
	go func() {
		defer wg.Done()

		err := m.Announce(ctx)
		if err != nil {
			errCh <- err
		}
	}()

	go func() {
		select {
		case <-ctx.Done():
		case err = <-errCh:
			shutdown()
		}
	}()

	wg.Wait()

	return nil
}

func (m *Module) Identities() <-chan id.Identity {
	m.mu.Lock()
	defer m.mu.Unlock()

	ch := make(chan id.Identity, len(m.entries))
	for hex := range m.entries {
		i, err := id.ParsePublicKeyHex(hex)
		if err != nil {
			panic(err)
		}
		ch <- i
	}
	close(ch)

	return ch
}

func (m *Module) Announce(ctx context.Context) error {
	return m.node.Infra().Announce(ctx)
}

func (m *Module) ignore(identity id.Identity) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.skip[identity.PublicKeyHex()] = struct{}{}
}

func (m *Module) unignore(identity id.Identity) {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.skip, identity.PublicKeyHex())
}

func (m *Module) handle(ctx context.Context, ip infra.Presence) {
	m.mu.Lock()
	defer m.mu.Unlock()

	hex := ip.Identity.PublicKeyHex()

	if e, found := m.entries[hex]; found {
		e.Touch()
		return
	}

	e := trackPresence(ctx, ip)
	m.entries[hex] = e

	// remove presence entry when it times out
	sig.OnCtx(ctx, e, func() {
		m.remove(hex)
	})

	log.Tag("presence").Info("%s present on %s", ip.Identity, log.Em(ip.Addr.Network()))

	m.events.Emit(EventIdentityPresent{ip.Identity, ip.Addr})

	_ = m.node.Tracker().Add(ip.Identity, ip.Addr, time.Now().Add(60*time.Minute))
}

func (m *Module) remove(hex string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	if e, found := m.entries[hex]; found {
		delete(m.entries, hex)

		log.Tag("presence").Info("%s gone from %s", e.id, log.Em(e.addr.Network()))

		m.events.Emit(EventIdentityGone{e.id})
	}
}
