package node

import (
	"context"
	"github.com/cryptopunkscc/astrald/auth/id"
	"io"
	"log"
)

func (node *Node) Query(ctx context.Context, remoteID id.Identity, query string) (io.ReadWriteCloser, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultQueryTimeout)
	defer cancel()

	log.Printf("[%s] -> %s\n", node.Contacts.DisplayName(remoteID), query)

	if remoteID.IsZero() || remoteID.IsEqual(node.identity) {
		return node.Ports.Query(query, node.identity)
	}

	go node.Linker.Link(ctx, remoteID)

	return node.Peers.Query(ctx, remoteID, query)
}
