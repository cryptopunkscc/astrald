package src

import (
	"encoding/hex"
	"strings"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/ops"
	"github.com/cryptopunkscc/astrald/mod/bip137sig"
)

type opEntropyToMnemonicArgs struct {
	Entropy string `query:"optional"`
	In      string `query:"optional"`
	Out     string `query:"optional"`
}

func (mod *Module) OpEntropyToMnemonic(
	ctx *astral.Context,
	q *ops.Query,
	args opEntropyToMnemonicArgs,
) (err error) {
	ch := channel.New(q.Accept(), channel.WithFormats(args.In, args.Out))
	defer ch.Close()

	if args.Entropy == "" {
		return ch.Send(astral.NewError("missing entropy"))
	}

	entropyBytes, err := hex.DecodeString(args.Entropy)
	if err != nil {
		return ch.Send(astral.NewError(err.Error()))
	}

	words, err := bip137sig.EntropyToMnemonic(entropyBytes)
	if err != nil {
		return ch.Send(astral.NewError(err.Error()))
	}

	phrase := strings.Join(words, " ")

	return ch.Send(astral.NewString16(phrase))
}
