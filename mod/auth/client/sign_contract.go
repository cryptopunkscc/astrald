package auth

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/query"
	"github.com/cryptopunkscc/astrald/mod/auth"
)

func (client *Client) SignContract(ctx *astral.Context, contract *auth.Contract) (signed *auth.SignedContract, err error) {
	ch, err := client.queryCh(ctx, auth.OpSignContract, nil)
	if err != nil {
		return
	}
	defer ch.Close()

	err = ch.Send(contract)
	if err != nil {
		return
	}

	err = ch.Switch(channel.Expect(&signed), channel.PassErrors)
	return
}

func SignContract(ctx *astral.Context, contract *auth.Contract) (*auth.SignedContract, error) {
	return Default().SignContract(ctx, contract)
}

func (client *Client) SignAsIssuer(ctx *astral.Context, contract *auth.Contract) (signed *auth.SignedContract, err error) {
	ch, err := client.queryCh(ctx, auth.OpSignContract, query.Args{"as": "issuer"})
	if err != nil {
		return
	}
	defer ch.Close()

	err = ch.Send(contract)
	if err != nil {
		return
	}

	err = ch.Switch(channel.Expect(&signed), channel.PassErrors)
	return
}

func SignAsIssuer(ctx *astral.Context, contract *auth.Contract) (*auth.SignedContract, error) {
	return Default().SignAsIssuer(ctx, contract)
}

func (client *Client) SignAsSubject(ctx *astral.Context, contract *auth.Contract) (signed *auth.SignedContract, err error) {
	ch, err := client.queryCh(ctx, auth.OpSignContract, query.Args{"as": "subject"})
	if err != nil {
		return
	}
	defer ch.Close()

	err = ch.Send(contract)
	if err != nil {
		return
	}

	err = ch.Switch(channel.Expect(&signed), channel.PassErrors)
	return
}

func SignAsSubject(ctx *astral.Context, contract *auth.Contract) (*auth.SignedContract, error) {
	return Default().SignAsSubject(ctx, contract)
}
