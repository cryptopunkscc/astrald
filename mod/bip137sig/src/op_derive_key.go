package src

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/ops"
	"github.com/cryptopunkscc/astrald/mod/bip137sig"
)

type opDeriveKeyArgs struct {
	Seed bip137sig.Seed
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

	privateKey, err := mod.DeriveKey(args.Seed, args.Path)
	if err != nil {
		return ch.Send(astral.Err(err))
	}

	return ch.Send(&privateKey)
}
