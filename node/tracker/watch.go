package tracker

import (
	"context"
	"github.com/cryptopunkscc/astrald/auth/id"
)

// Watch returns a channel which will receive all new addresses added to identity. The channel is closed when
// the context is done.
func (tracker *Tracker) Watch(ctx context.Context, nodeID id.Identity) <-chan *Addr {
	out := make(chan *Addr, 1)
	go func() {
		defer close(out)
		for event := range tracker.events.Subscribe(ctx) {
			if a, ok := event.(*EventNewAddr); ok {
				if a.NodeID.IsEqual(nodeID) {
					out <- a.Addr
				}
			}
		}
	}()
	return out
}
