package linkback

import (
	"context"
	"github.com/cryptopunkscc/astrald/node"
	"time"
)

const serviceHandle = ".linkback"
const ModuleName = "linkback"

type LinkBack struct{}

func (LinkBack) Run(ctx context.Context, n *node.Node) error {
	port, err := n.Ports.RegisterContext(ctx, serviceHandle)
	if err != nil {
		return err
	}

	go func() {
		for query := range port.Queries() {
			query.Accept().Close()
			n.Linking.Optimize(query.Caller(), 30*time.Second)
		}
	}()

	for event := range n.Follow(ctx) {
		if event, ok := event.(node.EventPeerLinked); ok {
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
