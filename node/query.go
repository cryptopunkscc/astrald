package node

import (
	"context"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/logfmt"
	"io"
	"log"
)

func (node *Node) Query(ctx context.Context, remoteID id.Identity, query string) (io.ReadWriteCloser, error) {
	if remoteID.IsZero() || remoteID.IsEqual(node.identity) {
		return node.Ports.Query(query, node.identity)
	}

	p := node.makePeer(remoteID)

	// set up a query context
	ctx, cancel := context.WithTimeout(ctx, defaultQueryTimeout)
	defer cancel()

	l, err := node.Linker.Connect(ctx, p)
	if err != nil {
		return nil, err
	}

	log.Printf("-> [%s]:%s (%s)\n", logfmt.ID(remoteID), query, l.Network())

	return l.Query(query)
}
