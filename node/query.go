package node

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/node/link"
	"github.com/cryptopunkscc/astrald/streams"
)

func (node *Node) Query(ctx context.Context, remoteID id.Identity, query string) (link.BasicConn, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultQueryTimeout)
	defer cancel()

	if remoteID.IsZero() || remoteID.IsEqual(node.identity) {
		return node.Ports.Query(ctx, query, nil)
	}

	link, err := node.Peers.Link(ctx, remoteID)
	if err != nil {
		return nil, err
	}

	return link.Query(ctx, query)
}

func (node *Node) onQuery(query *link.Query) error {
	select {
	case node.queryQueue <- query:
	default:
		log.Error("query dropped due to queue overflow: %s", query.Query())
		return errors.New("query queue overflow")
	}
	return nil
}

func (node *Node) peerQueryWorker(ctx context.Context) error {
	for {
		select {
		case query := <-node.queryQueue:
			log.Log("worker: start query %s", query.Query())
			ctx, _ := context.WithTimeout(ctx, defaultQueryTimeout)
			node.executeQuery(ctx, query)
			log.Log("worker: done query %s", query.Query())

		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (node *Node) executeQuery(ctx context.Context, query *link.Query) error {
	// Query a session with the service
	localConn, err := node.Ports.Query(ctx, query.Query(), query.Link())
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
