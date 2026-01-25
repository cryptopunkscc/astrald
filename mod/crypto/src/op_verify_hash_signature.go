package crypto

import (
	"encoding/hex"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/ops"
	"github.com/cryptopunkscc/astrald/mod/crypto"
	"github.com/cryptopunkscc/astrald/mod/secp256k1"
)

type opVerifyHashSignatureArgs struct {
	Hash string `query:"optional"`
	Key  string `query:"optional"`
	In   string `query:"optional"`
	Out  string `query:"optional"`
}

func (mod *Module) OpVerifyHashSignature(ctx *astral.Context, q *ops.Query, args opVerifyHashSignatureArgs) (err error) {
	ch := channel.New(q.Accept(), channel.WithFormats(args.In, args.Out))
	defer ch.Close()

	var publicKey *crypto.PublicKey
	var hash []byte

	// check hash argument
	if len(args.Hash) > 0 {
		hash, err = hex.DecodeString(args.Hash)
		if err != nil {
			return ch.Send(astral.NewError(err.Error()))
		}
	}

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
		switch msg := object.(type) {
		case *crypto.Signature:
			// check errors
			switch {
			case publicKey == nil:
				ch.Send(astral.NewError("missing public key"))
				return

			case hash == nil:
				ch.Send(astral.NewError("missing hash"))
				return
			}

			// verify signature
			err = mod.VerifyHashSignature(publicKey, msg, hash)
			if err != nil {
				ch.Send(astral.NewError(err.Error()))
				return
			}

			ch.Send(&astral.Ack{})

		case *crypto.PublicKey:
			publicKey = msg
			ch.Send(&astral.Ack{})

		case *crypto.Hash:
			hash = *msg
			ch.Send(&astral.Ack{})

		default:
			ch.Send(astral.NewErrUnexpectedObject(object))
			ch.Close()
		}
	})
}
