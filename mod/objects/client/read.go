package objects

import (
	"io"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/lib/query"
)

func (client *Client) Read(ctx *astral.Context, objectID *astral.ObjectID, offset, limit int64) (io.ReadCloser, error) {
	return client.query(ctx, "objects.read", query.Args{
		"id":     objectID,
		"offset": offset,
		"limit":  limit,
		"zone":   "dvn",
	})
}

func Read(ctx *astral.Context, objectID *astral.ObjectID, offset, limit int64) (io.ReadCloser, error) {
	return Default().Read(ctx, objectID, offset, limit)
}
