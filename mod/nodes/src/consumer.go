package nodes

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/lib/query"
)

type Consumer struct {
	*Module
	ProviderID *astral.Identity
}

func NewConsumer(module *Module, providerID *astral.Identity) *Consumer {
	return &Consumer{Module: module, ProviderID: providerID}
}

func (mod *Consumer) Relay(ctx *astral.Context, nonce astral.Nonce, caller *astral.Identity, target *astral.Identity) (err error) {
	q := query.New(
		mod.node.Identity(),
		mod.ProviderID,
		"nodes.relay",
		opRelayArgs{
			Nonce:     nonce,
			SetCaller: caller,
			SetTarget: target,
		},
	)

	conn, err := query.Route(ctx, mod.node, q)
	if err != nil {
		return
	}

	ch := astral.NewChannel(conn)
	defer ch.Close()

	response, err := ch.Read()
	if err != nil {
		return err
	}

	switch response := response.(type) {
	case *astral.ErrorMessage:
		return response

	case *astral.Ack:
		return nil

	default:
		return astral.NewError("unexpected response")
	}
}
