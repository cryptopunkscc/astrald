package nat

import (
	"errors"
	"io"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/query"
	"github.com/cryptopunkscc/astrald/mod/nat"
)

func (client *Client) NewTraversal(ctx *astral.Context, target string) (astral.Object, error) {
	ch, err := client.queryCh(ctx, nat.MethodNewTraversal, query.Args{
		"target": target,
	})
	if err != nil {
		return nil, err
	}
	defer ch.Close()

	obj, err := ch.Receive()
	if err != nil {
		return nil, err
	}

	if errMsg, ok := obj.(*astral.ErrorMessage); ok {
		return nil, errMsg
	}

	return obj, nil
}

func (client *Client) ListPairs(ctx *astral.Context, with string) ([]*nat.TraversedPortPair, error) {
	args := query.Args{}
	if with != "" {
		args["with"] = with
	}

	ch, err := client.queryCh(ctx, nat.MethodListPairs, args)
	if err != nil {
		return nil, err
	}
	defer ch.Close()

	var pairs []*nat.TraversedPortPair

	err = ch.Collect(func(msg astral.Object) error {
		switch msg := msg.(type) {
		case *nat.TraversedPortPair:
			pairs = append(pairs, msg)
		case *astral.EOS:
			return io.EOF
		case *astral.ErrorMessage:
			return msg
		default:
			return errors.New("unexpected response type: " + msg.ObjectType())
		}
		return nil
	})
	if errors.Is(err, io.EOF) {
		err = nil
	}

	return pairs, err
}

func (client *Client) SetEnabled(ctx *astral.Context, enabled bool) error {
	ch, err := client.queryCh(ctx, nat.MethodSetEnabled, query.Args{
		"arg": enabled,
	})
	if err != nil {
		return err
	}
	defer ch.Close()

	msg, err := ch.Receive()
	switch msg := msg.(type) {
	case *astral.Ack:
		return nil
	case nil:
		return err
	case *astral.ErrorMessage:
		return msg
	default:
		return errors.New("unexpected response type")
	}
}

func (client *Client) StartTraversal(ctx *astral.Context, target string) (astral.Object, error) {
	ch, err := client.queryCh(ctx, nat.MethodStartNatTraversal, query.Args{
		"target": target,
	})
	if err != nil {
		return nil, err
	}
	defer ch.Close()

	obj, err := ch.Receive()
	if err != nil {
		return nil, err
	}

	if errMsg, ok := obj.(*astral.ErrorMessage); ok {
		return nil, errMsg
	}

	return obj, nil
}

func (client *Client) StartTraversalCh(ctx *astral.Context, out string) (*channel.Channel, error) {
	return client.queryCh(ctx.IncludeZone(astral.ZoneNetwork), nat.MethodStartNatTraversal, query.Args{
		"out": out,
	})
}

func (client *Client) PairTakeCh(ctx *astral.Context, pair astral.Nonce, initiate bool) (*channel.Channel, error) {
	return client.queryCh(ctx.IncludeZone(astral.ZoneNetwork), nat.MethodPairTake, query.Args{
		"pair":     pair,
		"initiate": initiate,
	})
}

func NewTraversal(ctx *astral.Context, target string) (astral.Object, error) {
	return Default().NewTraversal(ctx, target)
}

func StartTraversal(ctx *astral.Context, target string) (astral.Object, error) {
	return Default().StartTraversal(ctx, target)
}

func ListPairs(ctx *astral.Context, with string) ([]*nat.TraversedPortPair, error) {
	return Default().ListPairs(ctx, with)
}

func SetEnabled(ctx *astral.Context, enabled bool) error {
	return Default().SetEnabled(ctx, enabled)
}
