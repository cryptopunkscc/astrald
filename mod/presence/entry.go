package presence

import (
	"context"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/net"
	"github.com/cryptopunkscc/astrald/sig"
	"sync"
	"time"
)

const presenceTimeout = 5 * time.Minute

type entry struct {
	id       id.Identity
	lastSeen time.Time
	addr     net.Endpoint
	closeCh  chan struct{}
	mu       sync.Mutex
}

func trackPresence(ctx context.Context, presence Presence) *entry {
	e := &entry{
		id:       presence.Identity,
		lastSeen: time.Now(),
		addr:     presence.Endpoint,
		closeCh:  make(chan struct{}),
	}

	sig.OnCtx(ctx, sig.Idle(ctx, e, presenceTimeout), func() {
		close(e.closeCh)
	})

	return e
}

func (e *entry) Idle() time.Duration {
	e.mu.Lock()
	defer e.mu.Unlock()

	return time.Now().Sub(e.lastSeen)
}

func (e *entry) Done() <-chan struct{} {
	e.mu.Lock()
	defer e.mu.Unlock()

	return e.closeCh
}

func (e *entry) Touch() {
	e.mu.Lock()
	defer e.mu.Unlock()

	e.lastSeen = time.Now()
}
