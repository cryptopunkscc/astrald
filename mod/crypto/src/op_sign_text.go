package crypto

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/ops"
	"github.com/cryptopunkscc/astrald/mod/crypto"
	"github.com/cryptopunkscc/astrald/mod/secp256k1"
)

type opSignTextArgs struct {
	Text   string `query:"optional"`
	Key    string `query:"optional"`
	Scheme string `query:"optional"`
	In     string `query:"optional"`
	Out    string `query:"optional"`
}

func (mod *Module) OpSignText(ctx *astral.Context, q *ops.Query, args opSignTextArgs) (err error) {
	ch := channel.New(q.Accept(), channel.WithFormats(args.In, args.Out))
	defer ch.Close()

	// set defaults
	if args.Scheme == "" {
		args.Scheme = "bip137"
	}

	var signerKey = secp256k1.FromIdentity(q.Caller())

	// check key argument
	if len(args.Key) > 0 {
		err = signerKey.UnmarshalText([]byte(args.Key))
		if err != nil {
			return ch.Send(astral.Err(err))
		}
	}

	signer, err := mod.TextSigner(signerKey, args.Scheme)
	if err != nil {
		return ch.Send(astral.Err(err))
	}

	// signAndSend signs the text and sends the signature to the channel
	var signAndSend = func(text string) error {
		sig, err := signer.SignText(ctx, text)
		if err != nil {
			return ch.Send(astral.Err(err))
		}
		return ch.Send(sig)
	}

	if len(args.Text) > 0 {
		return signAndSend(args.Text)
	}

	// process channel
	return ch.Switch(
		func(key *crypto.PublicKey) error {
			signerKey = key
			return ch.Send(&astral.Ack{})
		},
		func(text *astral.String8) error {
			return signAndSend(text.String())
		},
		func(text *astral.String16) error {
			return signAndSend(text.String())
		},
		channel.StopOnEOS,
	)
}
