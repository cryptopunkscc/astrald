package linkback

import (
	"context"
	"github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/node/peers"
	"time"
)

const (
	serviceHandle    = ".linkback"
	ModuleName       = "linkback"
	linkbackDuration = 30 * time.Second
)

type LinkBack struct{}

func (LinkBack) Run(ctx context.Context, n *node.Node) error {
	port, err := n.Ports.RegisterContext(ctx, serviceHandle)
	if err != nil {
		return err
	}

	go func() {
		for query := range port.Queries() {
			// reject local queries
			if query.IsLocal() {
				query.Reject()
				continue
			}

			if conn, err := query.Accept(); err == nil {
				conn.Close()
			}

			//TODO: start linkback
		}
	}()

	for event := range n.Subscribe(ctx.Done()) {
		if event, ok := event.(peers.EventLinked); ok {
			if event.Link.Outbound() {
				if c, err := n.Query(ctx, event.Peer.Identity(), serviceHandle); err == nil {
					c.Close()
				}
			}
		}
	}

	return nil
}

func (LinkBack) String() string {
	return ModuleName
}
