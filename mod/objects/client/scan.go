package objects

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/query"
)

func (client *Client) Scan(ctx *astral.Context, repo string, follow bool) (<-chan *astral.ObjectID, *error) {
	ch, err := client.queryCh(ctx, "objects.scan", query.Args{
		"repo":   repo,
		"follow": follow,
	})
	if err != nil {
		return nil, &err
	}

	out := make(chan *astral.ObjectID)

	var errPtr = new(error)

	go func() {
		defer close(out)
		defer ch.Close()

		*errPtr = ch.Switch(channel.Chan(out), channel.StopOnEOS, channel.PassErrors, channel.WithContext(ctx))
		if *errPtr != nil {
			return
		}

		if !follow {
			return
		}

		// send the separator
		select {
		case <-ctx.Done():
			return
		case out <- nil:
		}

		// handle updates
		*errPtr = ch.Switch(channel.Chan(out), channel.StopOnEOS, channel.PassErrors, channel.WithContext(ctx))
	}()

	return out, errPtr
}

func Scan(ctx *astral.Context, repo string, follow bool) (<-chan *astral.ObjectID, *error) {
	return Default().Scan(ctx, repo, follow)
}
