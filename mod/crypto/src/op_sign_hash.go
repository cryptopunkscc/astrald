package crypto

import (
	"encoding/hex"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/mod/crypto/secp256k1"
	"github.com/cryptopunkscc/astrald/mod/shell"
)

type opSignHashArgs struct {
	Hash   string
	Scheme string `query:"optional"`
	In     string `query:"optional"`
	Out    string `query:"optional"`
}

func (mod *Module) OpSignHash(ctx *astral.Context, q shell.Query, args opSignHashArgs) (err error) {
	ch := channel.New(q.Accept(), channel.WithFormats(args.In, args.Out))
	defer ch.Close()

	if args.Scheme == "" {
		args.Scheme = "asn1"
	}

	hash, err := hex.DecodeString(args.Hash)
	if err != nil {
		return ch.Send(astral.NewError(err.Error()))
	}

	signer, err := mod.HashSigner(secp256k1.PublicKeyFromIdentity(q.Caller()), args.Scheme)
	if err != nil {
		return ch.Send(astral.NewError(err.Error()))
	}

	sig, err := signer.SignHash(ctx, hash)
	if err != nil {
		return ch.Send(astral.NewError(err.Error()))
	}

	return ch.Send(sig)
}
