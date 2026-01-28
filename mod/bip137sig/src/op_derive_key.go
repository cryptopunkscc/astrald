package src

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/ops"
	"github.com/cryptopunkscc/astrald/mod/bip137sig"
)

type opDeriveKeyArgs struct {
	Path string
	In   string `query:"optional"`
	Out  string `query:"optional"`
}

func (mod *Module) OpDeriveKey(
	ctx *astral.Context,
	q *ops.Query,
	args opDeriveKeyArgs,
) (err error) {
	ch := channel.New(q.Accept(), channel.WithFormats(args.In, args.Out))
	defer ch.Close()

	return ch.Handle(ctx, func(object astral.Object) {
		switch object := object.(type) {
		case *bip137sig.Seed:
			privateKey, err := mod.DeriveKey(*object, args.Path)
			if err != nil {
				ch.Send(astral.Err(err))
				return
			}

			ch.Send(&privateKey)
		default:
			ch.Send(astral.NewErrUnexpectedObject(object))
		}
	})
}
