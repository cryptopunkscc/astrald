package src

import (
	"strings"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/ops"
	"github.com/cryptopunkscc/astrald/mod/bip137sig"
)

type opSeedArgs struct {
	Passphrase string `query:"optional"`
	In         string `query:"optional"`
	Out        string `query:"optional"`
}

func (mod *Module) OpSeed(
	ctx *astral.Context,
	q *ops.Query,
	args opSeedArgs,
) (err error) {
	ch := channel.New(q.Accept(), channel.WithFormats(args.In, args.Out))
	defer ch.Close()

	handle := func(mnemonic string) error {
		words := strings.Fields(mnemonic)

		seed, err := bip137sig.MnemonicToSeed(words, args.Passphrase)
		if err != nil {
			return ch.Send(astral.Err(err))
		}
		return ch.Send(&seed)
	}

	return ch.Switch(
		func(msg *astral.String16) error {
			return handle(string(*msg))
		},
		channel.PassErrors,
	)
}
