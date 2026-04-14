package auth

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/ops"
	"github.com/cryptopunkscc/astrald/mod/auth"
)

type opSignContractArgs struct {
	As  string `query:"optional"`
	In  string `query:"optional"`
	Out string `query:"optional"`
}

func (mod *Module) OpSignContract(ctx *astral.Context, q *ops.Query, args opSignContractArgs) error {
	ch := q.AcceptChannel(channel.WithFormats(args.In, args.Out))
	defer ch.Close()

	var signedContract *auth.SignedContract

	err := ch.Switch(
		func(c *auth.Contract) error {
			signedContract = &auth.SignedContract{Contract: c}
			return channel.ErrBreak
		},
		func(in *auth.SignedContract) error {
			signedContract = in
			return channel.ErrBreak
		},
		channel.PassErrors,
	)
	if err != nil {
		return err
	}

	// operate phase
	if args.As != "subject" {
		sig, err := mod.SignIssuer(ctx, signedContract.Contract)
		if err != nil {
			return ch.Send(astral.Err(err))
		}

		signedContract.IssuerSig = sig
	}

	if args.As != "issuer" {
		sig, err := mod.SignSubject(ctx, signedContract.Contract)
		if err != nil {
			return ch.Send(astral.Err(err))
		}
		signedContract.SubjecSig = sig
	}

	return ch.Send(signedContract)
}
