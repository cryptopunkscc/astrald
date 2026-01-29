package crypto

import (
	"encoding/hex"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/ops"
	"github.com/cryptopunkscc/astrald/mod/crypto"
	"github.com/cryptopunkscc/astrald/mod/secp256k1"
)

type opSignHashArgs struct {
	Hash   string
	Key    string `query:"optional"`
	Scheme string `query:"optional"`
	In     string `query:"optional"`
	Out    string `query:"optional"`
}

func (mod *Module) OpSignHash(ctx *astral.Context, q *ops.Query, args opSignHashArgs) (err error) {
	ch := channel.New(q.Accept(), channel.WithFormats(args.In, args.Out))
	defer ch.Close()

	if args.Scheme == "" {
		args.Scheme = "asn1"
	}

	signerKey := secp256k1.FromIdentity(q.Caller())

	// check key argument
	if len(args.Key) > 0 {
		signerKey = &crypto.PublicKey{}
		err = signerKey.UnmarshalText([]byte(args.Key))
		if err != nil {
			return ch.Send(astral.NewError(err.Error()))
		}
	}

	hash, err := hex.DecodeString(args.Hash)
	if err != nil {
		return ch.Send(astral.NewError(err.Error()))
	}

	signer, err := mod.HashSigner(signerKey, args.Scheme)
	if err != nil {
		return ch.Send(astral.NewError(err.Error()))
	}

	sig, err := signer.SignHash(ctx, hash)
	if err != nil {
		return ch.Send(astral.NewError(err.Error()))
	}

	return ch.Send(sig)
}
