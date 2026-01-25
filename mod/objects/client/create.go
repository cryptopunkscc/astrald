package objects

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/query"
	"github.com/cryptopunkscc/astrald/mod/objects"
)

func (client *Client) Create(ctx *astral.Context, repo string, alloc int) (objects.Writer, error) {
	// prepare arguments
	args := query.Args{}

	if alloc > 0 {
		args["alloc"] = alloc
	}
	if len(repo) > 0 {
		args["repo"] = repo
	}

	// send the query
	ch, err := client.queryCh(ctx, "objects.create", args)
	if err != nil {
		return nil, err
	}
	defer ch.Close()

	// wait for ack
	err = ch.Switch(channel.ExpectAck, channel.PassErrors, channel.WithContext(ctx))
	if err != nil {
		return nil, err
	}
	return &writer{ch: ch}, nil
}

func Create(ctx *astral.Context, repo string, alloc int) (objects.Writer, error) {
	return Default().Create(ctx, repo, alloc)
}
