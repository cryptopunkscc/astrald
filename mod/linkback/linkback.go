package linkback

import (
	"context"
	"github.com/cryptopunkscc/astrald/node"
	"github.com/cryptopunkscc/astrald/node/peer"
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

			n.Linking.Optimize(query.Link().RemoteIdentity(), linkbackDuration)
		}
	}()

	for event := range n.Subscribe(ctx.Done()) {
		if event, ok := event.(peer.EventLinked); ok {
			if event.Link.Outbound() {
				if c, err := event.Peer.Query(ctx, serviceHandle); err == nil {
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
