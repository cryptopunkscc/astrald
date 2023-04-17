package tracker

import (
	"context"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/node/event"
)

// Watch returns a channel which will receive all new addresses added to identity. The channel is closed when
// the context is done.
func (tracker *CoreTracker) Watch(ctx context.Context, nodeID id.Identity) <-chan *Addr {
	out := make(chan *Addr, 1)
	go func() {
		defer close(out)
		event.Handle(ctx, &tracker.events, func(e EventNewAddr) error {
			if e.NodeID.IsEqual(nodeID) {
				out <- e.Addr
			}
			return nil
		})
	}()
	return out
}
