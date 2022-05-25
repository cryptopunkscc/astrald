package presence

import (
	"context"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/infra"
	"github.com/cryptopunkscc/astrald/node/event"
	"github.com/cryptopunkscc/astrald/sig"
	"log"
	"sync"
)

type Presence struct {
	entries map[string]*entry
	mu      sync.Mutex
	events  event.Queue
	skip    map[string]struct{}
	net     infra.PresenceNet
}

func Run(ctx context.Context, net infra.PresenceNet, eventParent *event.Queue) *Presence {
	p := &Presence{
		entries: make(map[string]*entry),
		skip:    make(map[string]struct{}),
		net:     net,
	}
	p.events.SetParent(eventParent)
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

func (p *Presence) Subscribe(cancel sig.Signal) <-chan event.Event {
	return p.events.Subscribe(cancel)
}

func (p *Presence) Announce(ctx context.Context) error {
	return p.net.Announce(ctx)
}

func (p *Presence) process(ctx context.Context) {
	discover, err := p.net.Discover(ctx)
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

func (p *Presence) handle(ctx context.Context, ip infra.Presence) {
	p.mu.Lock()
	defer p.mu.Unlock()

	hex := ip.Identity.PublicKeyHex()

	if e, found := p.entries[hex]; found {
		e.Touch()
		return
	}

	e := trackPresence(ctx, ip)
	p.entries[hex] = e

	// remove presence entry when it times out
	sig.OnCtx(ctx, e, func() {
		p.remove(hex)
	})

	p.events.Emit(EventIdentityPresent{ip.Identity, ip.Addr})
}

func (p *Presence) remove(hex string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if e, found := p.entries[hex]; found {
		delete(p.entries, hex)

		p.events.Emit(EventIdentityGone{e.id})
	}
}
