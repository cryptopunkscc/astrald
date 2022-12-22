package node

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/node/link"
	"github.com/cryptopunkscc/astrald/streams"
	"io"
)

func (node *Node) Query(ctx context.Context, remoteID id.Identity, query string) (io.ReadWriteCloser, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultQueryTimeout)
	defer cancel()

	if remoteID.IsZero() || remoteID.IsEqual(node.identity) {
		return node.Ports.Query(ctx, query, nil)
	}

	peer, err := node.Peers.Link(ctx, remoteID)
	if err != nil {
		return nil, err
	}

	link := link.Select(peer.Links(), link.LowestRoundTrip)
	if link == nil {
		return nil, errors.New("no viable link")
	}

	return link.Query(ctx, query)
}

func (node *Node) peerQueryWorker(ctx context.Context) error {
	for {
		select {
		case query := <-node.Peers.Queries():
			ctx, _ := context.WithTimeout(ctx, defaultQueryTimeout)
			node.executeQuery(ctx, query)

		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (node *Node) executeQuery(ctx context.Context, query *link.Query) error {
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
