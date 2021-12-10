package presence

import (
	"context"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/infra"
	"github.com/cryptopunkscc/astrald/sig"
	"log"
	"sync"
)

type Presence struct {
	entries map[string]*entry
	mu      sync.Mutex
	events  chan Event
	skip    map[string]struct{}
}

func Run(ctx context.Context) *Presence {
	p := &Presence{
		entries: make(map[string]*entry),
		events:  make(chan Event),
		skip:    make(map[string]struct{}),
	}
	go p.process(ctx)
	return p
}

func (p *Presence) Identities() <-chan id.Identity {
	p.mu.Lock()
	defer p.mu.Unlock()

	ch := make(chan id.Identity, len(p.entries))
	for hex := range p.entries {
		i, err := id.ParsePublicKeyHex(hex)
		if err != nil {
			panic(err)
		}
		ch <- i
	}
	close(ch)

	return ch
}

func (p *Presence) Events() <-chan Event {
	return p.events
}

func (p *Presence) Announce(ctx context.Context, identity id.Identity) error {
	p.ignore(identity)

	err := astral.Announce(ctx, identity)
	if err != nil {
		p.unignore(identity)
		return err
	}

	go func() {
		<-ctx.Done()
		p.unignore(identity)
	}()

	return nil
}

func (p *Presence) process(ctx context.Context) {
	discover, err := astral.Discover(ctx)
	if err != nil {
		log.Println("discover error:", err)
		return
	}

	for presence := range discover {
		hex := presence.Identity.PublicKeyHex()
		if _, skip := p.skip[hex]; skip {
			continue
		}

		p.handle(ctx, presence)
	}
}

func (p *Presence) ignore(identity id.Identity) {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.skip[identity.PublicKeyHex()] = struct{}{}
}

func (p *Presence) unignore(identity id.Identity) {
	p.mu.Lock()
	defer p.mu.Unlock()

	delete(p.skip, identity.PublicKeyHex())
}

func (p *Presence) handle(ctx context.Context, presence infra.Presence) {
	p.mu.Lock()
	defer p.mu.Unlock()

	hex := presence.Identity.PublicKeyHex()

	if e, found := p.entries[hex]; found {
		e.Touch()
		return
	}

	e := trackPresence(ctx, presence)
	p.entries[hex] = e

	// remove presence entry when it times out
	sig.On(ctx, e, func() {
		p.remove(hex)
	})

	p.events <- Event{
		identity: presence.Identity,
		event:    EventIdentityPresent,
		addr:     presence.Addr,
	}
}

func (p *Presence) remove(hex string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if e, found := p.entries[hex]; found {
		delete(p.entries, hex)

		p.events <- Event{
			identity: e.id,
			event:    EventIdentityGone,
		}
	}
}
