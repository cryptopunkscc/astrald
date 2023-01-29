package infra

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/node/link"
)

type FilteredQuerier struct {
	Querier
	FilteredID id.Identity
}

func (q FilteredQuerier) Query(ctx context.Context, remoteID id.Identity, query string) (link.BasicConn, error) {
	if remoteID.IsEqual(q.FilteredID) {
		return nil, errors.New("filtered")
	}
	return q.Querier.Query(ctx, remoteID, query)
}
