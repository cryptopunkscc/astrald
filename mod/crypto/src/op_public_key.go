package crypto

import (
	"io"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/ops"
	"github.com/cryptopunkscc/astrald/mod/crypto"
	"github.com/cryptopunkscc/astrald/mod/secp256k1"
)

type opPublicKeyArgs struct {
	In  string `query:"optional"`
	Out string `query:"optional"`
}

func (mod *Module) OpPublicKey(ctx *astral.Context, q *ops.Query, args opPublicKeyArgs) (err error) {
	ch := channel.New(q.Accept(), channel.WithFormats(args.In, args.Out))
	defer ch.Close()

	err = ch.Collect(func(msg astral.Object) (err error) {
		switch msg := msg.(type) {
		case *crypto.PrivateKey:
			err = ch.Send(secp256k1.PublicKey(msg))

		case *astral.EOS:
			return io.EOF

		default:
			err = ch.Send(astral.NewErrUnexpectedObject(msg))
		}

		return
	})

	return err
}
