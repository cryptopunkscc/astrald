package node

import (
	"context"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/node/link"
	"io"
	"log"
	"strings"
	"time"
)

const logSilent = false

func (node *Node) Query(ctx context.Context, remoteID id.Identity, query string) (io.ReadWriteCloser, error) {
	ctx, cancel := context.WithTimeout(ctx, defaultQueryTimeout)
	defer cancel()

	if !strings.HasPrefix(query, ".") || logSilent {
		log.Printf("[%s] -> %s\n", node.Contacts.DisplayName(remoteID), query)
	}

	if remoteID.IsZero() || remoteID.IsEqual(node.identity) {
		return node.Ports.Query(query, node.identity)
	}

	node.Linking.Optimize(remoteID, 30*time.Second)

	return node.Peers.Query(ctx, remoteID, query)
}

func (node *Node) processQueries(ctx context.Context) {
	for {
		select {
		case query := <-node.queries:
			node.handleQuery(query)
		case <-ctx.Done():
			return
		}
	}
}

func (node *Node) handleQuery(query *link.Query) error {
	// log non-silent queries
	if !query.IsSilent() || logSilent {
		log.Printf("[%s] <- %s (%s)\n", node.Contacts.DisplayName(query.Caller()), query, query.Link().Network())
	}

	// Query a session with the service
	localStream, err := node.Ports.Query(query.String(), query.Caller())
	if err != nil {
		query.Reject()
		return err
	}

	// Accept remote party's query
	remoteStream, err := query.Accept()
	if err != nil {
		localStream.Close()
		return err
	}

	// Connect local and remote streams
	go func() {
		_, _ = io.Copy(localStream, remoteStream)
		_ = localStream.Close()
	}()
	go func() {
		_, _ = io.Copy(remoteStream, localStream)
		_ = remoteStream.Close()
	}()

	return nil
}
