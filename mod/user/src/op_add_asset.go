package user

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/ops"
)

type opAddAssetArgs struct {
	ID  *astral.ObjectID
	Out string `query:"optional"`
}

func (mod *Module) OpAddAsset(ctx *astral.Context, q *ops.Query, args opAddAssetArgs) (err error) {
	err = mod.AddAsset(args.ID)
	if err != nil {
		mod.log.Error("error adding asset: %v", err)
		return q.RejectWithCode(astral.CodeInternalError)
	}

	ch := channel.New(q.Accept(), channel.WithOutputFormat(args.Out))
	defer ch.Close()

	return ch.Send(&astral.Ack{})
}
