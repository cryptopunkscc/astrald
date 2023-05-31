package keepalive

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/mod/contacts"
	"github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/node/link"
	"github.com/cryptopunkscc/astrald/node/modules"
	"github.com/cryptopunkscc/astrald/node/services"
	"sync"
	"time"
)

const portName = "net.keepalive"

type Module struct {
	node   node.Node
	config Config
}

// time between successive link retries, in seconds
var relinkIntervals = []int{5, 5, 15, 30, 60, 60, 60, 60, 5 * 60, 5 * 60, 5 * 60, 5 * 60, 15 * 60}

// interval between periodic checks for new best link
const checkBestLinkInterval = 5 * time.Minute

func (m *Module) Run(ctx context.Context) error {
	_, err := modules.WaitReady[*contacts.Module](ctx, m.node.Modules())
	if err != nil {
		return err
	}

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()

		if err := m.runServer(ctx); err != nil {
			log.Errorv(1, "error running server: %s", err)
		}
	}()

	for _, sn := range m.config.StickyNodes {
		nodeID, err := m.node.Resolver().Resolve(sn)
		if err != nil {
			log.Error("error resolving %s: %s", sn, err)
			continue
		}

		wg.Add(1)
		go func() {
			defer wg.Done()
			m.keepNodeLinked(ctx, nodeID)
		}()
	}

	wg.Wait()
	return nil
}

func (m *Module) runServer(ctx context.Context) error {
	_, err := m.node.Services().Register(ctx, m.node.Identity(), portName, func(query *services.Query) error {
		m.handleQuery(query)
		return nil
	})
	return err
}

func (m *Module) handleQuery(q *services.Query) error {
	if q.Source() == services.SourceLocal {
		q.Reject()
		return errors.New("local query not allowed")
	}

	conn, err := q.Accept()
	if err == nil {
		conn.Close()
	}

	// disable timeout on the link
	conn.Link().Idle().SetTimeout(0)
	log.Log("timeout disabled for %s over %s",
		conn.Link().RemoteIdentity(),
		conn.Link().Network(),
	)

	return nil
}

func (m *Module) keepNodeLinked(ctx context.Context, nodeID id.Identity) error {
	var errc int
	var best *link.Link

	newLinksCh := m.subscribeNewLinksWithNode(ctx, nodeID)

	log.Logv(1, "will keep %s linked", nodeID)

	for {
		newBest, err := m.node.Network().Link(ctx, nodeID)

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
			if conn, err := best.Query(ctx, "net.keepalive"); err == nil {
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

func (m *Module) subscribeNewLinksWithNode(ctx context.Context, nodeID id.Identity) <-chan *link.Link {
	var ch = make(chan *link.Link)

	go func() {
		defer close(ch)
		for event := range m.node.Network().Events().Subscribe(ctx) {
			if event, ok := event.(link.EventLinkEstablished); ok {
				if event.Link.RemoteIdentity().IsEqual(nodeID) {
					ch <- event.Link
				}
			}
		}
	}()

	return ch
}
