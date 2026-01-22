package crypto

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/mod/crypto"
	"github.com/cryptopunkscc/astrald/mod/secp256k1"
	"github.com/cryptopunkscc/astrald/mod/shell"
)

type opSignMessageArgs struct {
	Key    string `query:"optional"`
	Scheme string `query:"optional"`
	In     string `query:"optional"`
	Out    string `query:"optional"`
}

func (mod *Module) OpSignMessage(ctx *astral.Context, q shell.Query, args opSignMessageArgs) (err error) {
	ch := channel.New(q.Accept(), channel.WithFormats(args.In, args.Out))
	defer ch.Close()

	if args.Scheme == "" {
		args.Scheme = "bip137"
	}

	var publicKey *crypto.PublicKey

	// check key argument
	if len(args.Key) > 0 {
		publicKey = &crypto.PublicKey{}
		err = publicKey.UnmarshalText([]byte(args.Key))
		if err != nil {
			return ch.Send(astral.NewError(err.Error()))
		}
	} else {
		publicKey = secp256k1.FromIdentity(q.Caller())
	}

	signer, err := mod.MessageSigner(publicKey, args.Scheme)
	if err != nil {
		return ch.Send(astral.NewError(err.Error()))
	}

	// process channel
	return ch.Handle(ctx, func(object astral.Object) {
		switch object := object.(type) {
		case *crypto.PublicKey:
			publicKey = object
			ch.Send(&astral.Ack{})

		case *astral.String8:
			msg := (string)(*object)

			sig, err := signer.SignMessage(ctx, msg)
			if err != nil {
				ch.Send(astral.NewError(err.Error()))
				return
			}

			ch.Send(sig)

		default:
			ch.Send(astral.NewErrUnexpectedObject(object))
			ch.Close()
		}
	})
}
