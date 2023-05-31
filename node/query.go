package node

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/node/link"
	"github.com/cryptopunkscc/astrald/streams"
	"time"
)

func (node *CoreNode) Query(ctx context.Context, remoteID id.Identity, query string) (link.BasicConn, error) {
	if remoteID.IsZero() || remoteID.IsEqual(node.identity) {
		return node.Services().Query(ctx, node.identity, query, nil)
	}

	return node.Network().Query(ctx, remoteID, query)
}

func (node *CoreNode) onQuery(query *link.Query) error {
	select {
	case node.queryQueue <- query:
	default:
		node.log.Error("query dropped due to queue overflow: %s", query.Query())
		return errors.New("query queue overflow")
	}
	return nil
}

func (node *CoreNode) peerQueryWorker(ctx context.Context) error {
	for {
		select {
		case query := <-node.queryQueue:
			ctx, _ := context.WithTimeout(ctx, defaultQueryTimeout)
			var start = time.Now()
			var err = node.executeQuery(ctx, query)
			var elapsed = time.Since(start)

			node.log.Logv(2, "served query %s for %s (time %s, err %s)",
				query.Query(),
				query.Link().RemoteIdentity(),
				elapsed.Round(time.Microsecond),
				err,
			)

		case <-ctx.Done():
			return ctx.Err()
		}
	}
}

func (node *CoreNode) executeQuery(ctx context.Context, query *link.Query) error {
	// Query a session with the service
	localConn, err := node.Services().Query(ctx, query.Link().RemoteIdentity(), query.Query(), query.Link())
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
