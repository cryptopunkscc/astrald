package user

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/ops"
	"github.com/cryptopunkscc/astrald/mod/user"
)

type opAddToIndexArgs struct {
	In  string `query:"optional"`
	Out string `query:"optional"`
}

func (mod *Module) OpAddToIndex(ctx *astral.Context, query *ops.Query, args opAddToIndexArgs) (err error) {
	ch := query.AcceptChannel(channel.WithFormats(args.In, args.Out))
	defer ch.Close()

	return ch.Switch(
		func(contract *user.SignedNodeContract) error {
			err := mod.IndexSignedNodeContract(contract)
			if err != nil {
				return ch.Send(astral.Err(err))
			}
			return ch.Send(&astral.Ack{})
		},
		channel.StopOnEOS,
	)
}
