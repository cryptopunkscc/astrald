package node

import (
	"context"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/node/link"
	"github.com/cryptopunkscc/astrald/streams"
	"io"
)

func (node *Node) Query(ctx context.Context, remoteID id.Identity, query string) (io.ReadWriteCloser, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultQueryTimeout)
	defer cancel()

	//TODO: Emit an event for logging?
	//if !strings.HasPrefix(query, ".") || logSilent {
	//	log.Printf("[%s] -> %s\n", node.Contacts.DisplayName(remoteID), query)
	//}

	if remoteID.IsZero() || remoteID.IsEqual(node.identity) {
		return node.Ports.Query(ctx, query, nil)
	}

	peer, err := node.Peers.Link(ctx, remoteID)
	if err != nil {
		return nil, err
	}

	link := link.Select(peer.Links(), link.LowestRoundTrip)

	return link.Query(ctx, query)
}

func (node *Node) processQueries(ctx context.Context) {
	for {
		select {
		case query := <-node.Peers.Queries():
			ctx, _ := context.WithTimeout(ctx, defaultQueryTimeout)
			node.handleQuery(ctx, query)
		case <-ctx.Done():
			return
		}
	}
}

func (node *Node) handleQuery(ctx context.Context, query *link.Query) error {
	//TODO: Emit an event for logging?
	//if !query.IsSilent() || logSilent {
	//	  log.Printf("[%s] <- %s (%s)\n", node.Contacts.DisplayName(query.Caller()), query, query.Link().Network())
	//}

	// Query a session with the service
	localConn, err := node.Ports.Query(ctx, query.String(), query.Link())
	if err != nil {
		query.Reject()
		return err
	}

	// Accept remote party's query
	remoteConn, err := query.Accept()
	if err != nil {
		localConn.Close()
		return err
	}

	go streams.Join(localConn, remoteConn)

	return nil
}
