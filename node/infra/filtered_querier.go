package infra

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/auth/id"
	"io"
)

type FilteredQuerier struct {
	Querier
	FilteredID id.Identity
}

func (q FilteredQuerier) Query(ctx context.Context, remoteID id.Identity, query string) (io.ReadWriteCloser, error) {
	if remoteID.IsEqual(q.FilteredID) {
		return nil, errors.New("filtered")
	}
	return q.Querier.Query(ctx, remoteID, query)
}
