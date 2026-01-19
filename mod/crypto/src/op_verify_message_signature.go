package crypto

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/mod/crypto"
	"github.com/cryptopunkscc/astrald/mod/secp256k1"
	"github.com/cryptopunkscc/astrald/mod/shell"
)

type opVerifyMessageSignatureArgs struct {
	Msg string `query:"optional"`
	Key string `query:"optional"`
	In  string `query:"optional"`
	Out string `query:"optional"`
}

func (mod *Module) OpVerifyMessageSignature(ctx *astral.Context, q shell.Query, args opVerifyMessageSignatureArgs) (err error) {
	ch := channel.New(q.Accept(), channel.WithFormats(args.In, args.Out))
	defer ch.Close()

	var publicKey *crypto.PublicKey
	var msg = args.Msg

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

	// process channel
	return ch.Handle(ctx, func(object astral.Object) {
		switch object := object.(type) {
		case *crypto.Signature:
			// check errors
			switch {
			case publicKey == nil:
				ch.Send(astral.NewError("missing public key"))
				return

			case msg == "":
				ch.Send(astral.NewError("missing message"))
				return
			}

			// verify signature
			err = mod.VerifyMessageSignature(publicKey, object, msg)
			if err != nil {
				ch.Send(astral.NewError(err.Error()))
				return
			}

			ch.Send(&astral.Ack{})

		case *crypto.PublicKey:
			publicKey = object
			ch.Send(&astral.Ack{})

		case *astral.String8:
			msg = (string)(*object)
			ch.Send(&astral.Ack{})

		default:
			ch.Send(astral.NewError("unexpected object type"))
			ch.Close()
		}
	})
}
