package objects

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/query"
	"github.com/cryptopunkscc/astrald/mod/objects"
)

func (client *Client) Search(ctx *astral.Context, q string) (<-chan *objects.SearchResult, *error) {
	ch, err := client.queryCh(ctx, "objects.search", query.Args{
		"q": q,
	})
	if err != nil {
		return nil, &err
	}

	var out = make(chan *objects.SearchResult)
	var errPtr = new(error)

	go func() {
		defer close(out)
		defer ch.Close()

		*errPtr = ch.Switch(channel.Chan(out), channel.StopOnEOS, channel.PassErrors, channel.WithContext(ctx))
	}()

	return out, errPtr
}

func Search(ctx *astral.Context, q string) (<-chan *objects.SearchResult, *error) {
	return Default().Search(ctx, q)
}
