package nodes

import (
	"github.com/cryptopunkscc/astrald/astral"
	nodesmod "github.com/cryptopunkscc/astrald/mod/nodes"
)

func (client *Client) SendRelayedQuery(ctx *astral.Context, container *nodesmod.QueryContainer) error {
	ch, err := client.queryCh(ctx, nodesmod.MethodNodeOpenRelay, nil)
	if err != nil {
		return err
	}
	defer ch.Close()

	if err := ch.Send(container); err != nil {
		return err
	}

	response, err := ch.Receive()
	if err != nil {
		return err
	}

	switch r := response.(type) {
	case *astral.ErrorMessage:
		return r
	case *astral.Ack:
		return nil
	default:
		return astral.NewErrUnexpectedObject(r)
	}
}
