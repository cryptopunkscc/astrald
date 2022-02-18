package node

import (
	"context"
	"errors"
	alink "github.com/cryptopunkscc/astrald/link"
	nlink "github.com/cryptopunkscc/astrald/node/link"
	"log"
)

func (node *Node) AddLink(link *alink.Link) error {
	if link == nil {
		return errors.New("link is nil")
	}

	node.links <- link

	return nil
}

func (node *Node) processLinks(ctx context.Context) {
	for {
		select {
		case link := <-node.links:
			if err := node.addLink(ctx, nlink.Wrap(link, &node.events)); err != nil {
				log.Println("[node] link processing error:", err)
			}
		case <-ctx.Done():
			return
		}
	}
}

func (node *Node) addLink(ctx context.Context, link *nlink.Link) error {
	if err := node.Peers.Add(link); err != nil {
		return err
	}

	// forward queries coming from the link
	go func() {
		for query := range link.Queries() {
			node.queries <- query
		}
	}()

	return nil
}
