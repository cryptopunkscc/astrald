package src

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/ops"
	"github.com/cryptopunkscc/astrald/mod/bip137sig"
)

type opGenerateEntropyArgs struct {
	Bits int    `query:"optional"`
	In   string `query:"optional"`
	Out  string `query:"optional"`
}

func (mod *Module) OpGenerateEntropy(
	ctx *astral.Context,
	q *ops.Query,
	args opGenerateEntropyArgs,
) (err error) {
	ch := channel.New(q.Accept(), channel.WithFormats(args.In, args.Out))
	defer ch.Close()

	bits := args.Bits
	if bits == 0 {
		bits = 128
	}

	entropy, err := bip137sig.NewEntropy(bits)
	if err != nil {
		return ch.Send(astral.NewError(err.Error()))
	}

	return ch.Send(bip137sig.Entropy{Data: entropy})
}
