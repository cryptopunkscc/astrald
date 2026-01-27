package src

import (
	"fmt"
	"strings"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/ops"
	"github.com/cryptopunkscc/astrald/mod/bip137sig"
)

type opMnemonicToSeedArgs struct {
	Passphrase string `query:"optional"`
	In         string `query:"optional"`
	Out        string `query:"optional"`
}

func (mod *Module) OpMnemonicToSeed(
	ctx *astral.Context,
	q *ops.Query,
	args opMnemonicToSeedArgs,
) (err error) {
	ch := channel.New(q.Accept(), channel.WithFormats(args.In, args.Out))
	defer ch.Close()

	fmt.Println("TEST TEST")

	handle := func(mnemonic string) error {
		words := strings.Fields(mnemonic)
		if len(words) == 0 {
			return ch.Send(astral.Err(bip137sig.ErrInvalidMnemonic))
		}

		if _, err := bip137sig.MnemonicToEntropy(words); err != nil {
			return ch.Send(astral.Err(err))
		}

		seed := bip137sig.MnemonicToSeed(words, args.Passphrase)
		return ch.Send(&seed)
	}

	return ch.Switch(
		func(msg *astral.String16) error {
			return handle(string(*msg))
		},
		channel.PassErrors,
	)
}
