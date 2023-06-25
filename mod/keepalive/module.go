package keepalive

import (
	"context"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/node"
	"sync"
	"time"
)

type Module struct {
	config Config
	node   node.Node
	log    *log.Logger
	ctx    context.Context
}

// time between successive link retries, in seconds
var relinkIntervals = []int{5, 5, 15, 30, 60, 60, 60, 60, 60, 60, 60, 60, 60, 5 * 60, 5 * 60, 5 * 60, 5 * 60, 15 * 60}

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

	wg.Wait()

	return nil
}

func (module *Module) keepNodeLinked(ctx context.Context, nodeID id.Identity) error {
	var errc int

	module.log.Logv(1, "will keep %s linked", nodeID)

	for {
		best, err := module.node.Network().Link(ctx, nodeID)

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
	wait:
		for {
			select {
			case <-best.Done(): // find new best when current best closes
				break wait

			case <-ctx.Done(): // abort when context ends
				return ctx.Err()
			}
		}
	}
}
