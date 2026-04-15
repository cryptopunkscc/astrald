package user

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/mod/auth"
	"github.com/cryptopunkscc/astrald/mod/crypto"
	"github.com/cryptopunkscc/astrald/mod/user"
)

func (client *Client) Invite(ctx *astral.Context, contract *auth.Contract, issuerSig *crypto.Signature) (subjectSig *crypto.Signature, err error) {
	ch, err := client.queryCh(ctx, user.OpInvite, nil)
	if err != nil {
		return
	}
	defer ch.Close()

	err = ch.Send(contract)
	if err != nil {
		return
	}

	err = ch.Send(issuerSig)
	if err != nil {
		return
	}

	err = ch.Switch(channel.Expect(&subjectSig), channel.PassErrors)
	return
}
