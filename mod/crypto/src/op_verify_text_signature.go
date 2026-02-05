package crypto

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/ops"
	"github.com/cryptopunkscc/astrald/mod/crypto"
	"github.com/cryptopunkscc/astrald/mod/secp256k1"
)

type opVerifyTextSignatureArgs struct {
	Text string `query:"optional"`
	Key  string `query:"optional"`
	In   string `query:"optional"`
	Out  string `query:"optional"`
}

func (mod *Module) OpVerifyTextSignature(ctx *astral.Context, q *ops.Query, args opVerifyTextSignatureArgs) (err error) {
	ch := channel.New(q.Accept(), channel.WithFormats(args.In, args.Out))
	defer ch.Close()

	var publicKey = secp256k1.FromIdentity(q.Caller())
	var text = args.Text

	// check key argument
	if len(args.Key) > 0 {
		err = publicKey.UnmarshalText([]byte(args.Key))
		if err != nil {
			return ch.Send(astral.Err(err))
		}
	}

	return ch.Switch(
		func(key *crypto.PublicKey) error {
			publicKey = key
			return ch.Send(&astral.Ack{})
		},
		func(text2 *astral.String8) error {
			text = (string)(*text2)
			return ch.Send(&astral.Ack{})
		},
		func(text2 *astral.String16) error {
			text = (string)(*text2)
			return ch.Send(&astral.Ack{})
		},
		func(sig *crypto.Signature) error {
			switch {
			case publicKey == nil:
				return ch.Send(astral.NewError("missing public key"))
			case text == "":
				return ch.Send(astral.NewError("missing text"))
			}

			err = mod.VerifyTextSignature(publicKey, sig, text)
			if err != nil {
				return ch.Send(astral.Err(err))
			}
			return ch.Send(&astral.Ack{})
		},
		channel.StopOnEOS,
	)
}
