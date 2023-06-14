package keepalive

import (
	"context"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/node/link"
	"github.com/cryptopunkscc/astrald/query"
	"sync"
	"time"
)

const serviceName = "net.keepalive"

type Module struct {
	config Config
	node   node.Node
	log    *log.Logger
	ctx    context.Context
}

// time between successive link retries, in seconds
var relinkIntervals = []int{5, 5, 15, 30, 60, 60, 60, 60, 5 * 60, 5 * 60, 5 * 60, 5 * 60, 15 * 60}

// interval between periodic checks for new best link
const checkBestLinkInterval = 5 * time.Minute

func (module *Module) Run(ctx context.Context) error {
	module.ctx = ctx
	var wg sync.WaitGroup

	for _, sn := range module.config.StickyNodes {
		nodeID, err := module.node.Resolver().Resolve(sn)
		if err != nil {
			module.log.Error("error resolving %s: %s", sn, err)
			continue
		}

		wg.Add(1)
		go func() {
			defer wg.Done()
			module.keepNodeLinked(ctx, nodeID)
		}()
	}

	return NewServer(module).Run(ctx)
}

func (module *Module) keepNodeLinked(ctx context.Context, nodeID id.Identity) error {
	var errc int
	var best *link.Link

	newLinksCh := module.subscribeNewLinksWithNode(ctx, nodeID)

	module.log.Logv(1, "will keep %s linked", nodeID)

	for {
		newBest, err := module.node.Network().Link(ctx, nodeID)

		if err != nil {
			var ival time.Duration
			if errc < len(relinkIntervals) {
				ival = time.Duration(relinkIntervals[errc]) * time.Second
			} else {
				ival = time.Duration(relinkIntervals[len(relinkIntervals)-1]) * time.Second
			}

			select {
			case <-time.After(ival):
				errc++
				continue

			case <-ctx.Done():
				return ctx.Err()
			}
		}

		errc = 0

		if best != newBest {
			// keep the best link always active so that it never times out
			if best != nil {
				best.Activity().Done()
			}
			best = newBest
			best.Activity().Add(1)

			// ask the other node to never close the link due to inactivity
			conn, err := query.Run(ctx,
				best,
				query.New(module.node.Identity(), best.RemoteIdentity(), serviceName),
			)
			if err == nil {
				conn.Close()
			}
		}

	wait:
		for {
			select {
			case <-best.Done(): // find new best when current best closes
				break wait

			case <-newLinksCh: // find new best when a new link was established
				break wait

			case <-time.After(checkBestLinkInterval): // check periodically for quality (latency) changes within links
				break wait

			case <-ctx.Done(): // abort when context ends
				return ctx.Err()
			}
		}
	}
}

func (module *Module) subscribeNewLinksWithNode(ctx context.Context, nodeID id.Identity) <-chan *link.Link {
	var ch = make(chan *link.Link)

	go func() {
		defer close(ch)
		for event := range module.node.Network().Events().Subscribe(ctx) {
			if event, ok := event.(link.EventLinkEstablished); ok {
				if event.Link.RemoteIdentity().IsEqual(nodeID) {
					ch <- event.Link
				}
			}
		}
	}()

	return ch
}
