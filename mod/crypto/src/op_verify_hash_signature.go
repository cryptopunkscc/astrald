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
			return ch.Send(astral.Err(err))
		}
	}

	// check key argument
	if len(args.Key) > 0 {
		publicKey = &crypto.PublicKey{}
		err = publicKey.UnmarshalText([]byte(args.Key))
		if err != nil {
			return ch.Send(astral.Err(err))
		}
	} else {
		publicKey = secp256k1.FromIdentity(q.Caller())
	}

	// process the channel
	return ch.Switch(
		func(sig *crypto.Signature) error {
			// check errors
			switch {
			case publicKey == nil:
				return ch.Send(astral.NewError("missing public key"))
			case hash == nil:
				return ch.Send(astral.NewError("missing hash"))
			}

			// verify signature
			err = mod.VerifyHashSignature(publicKey, sig, hash)
			if err != nil {
				return ch.Send(astral.Err(err))
			}
			return ch.Send(&astral.Ack{})
		},
		func(key *crypto.PublicKey) error {
			publicKey = key
			return ch.Send(&astral.Ack{})
		},
		func(hash2 *crypto.Hash) error {
			hash = *hash2
			return ch.Send(&astral.Ack{})
		},
		channel.BreakOnEOS,
	)
}
