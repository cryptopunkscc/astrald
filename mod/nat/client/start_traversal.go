package nat

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/query"
	"github.com/cryptopunkscc/astrald/mod/ip"
	"github.com/cryptopunkscc/astrald/mod/nat"
)

type TraversalClient struct {
	client  *Client
	localID *astral.Identity
	localIP ip.IP
	puncher nat.Puncher
}

func (client *Client) NewTraversalClient(localIP ip.IP, puncher nat.Puncher) *TraversalClient {
	return &TraversalClient{
		client:  client,
		localID: client.astral.GuestID(),
		localIP: localIP,
		puncher: puncher,
	}
}

func (t *TraversalClient) StartTraversal(ctx *astral.Context, target *astral.Identity) (*nat.TraversedPortPair, error) {
	ch, err := t.client.queryCh(ctx, nat.MethodStartNatTraversal, query.Args{
		"target": target.String(),
	})
	if err != nil {
		return nil, err
	}
	defer ch.Close()

	traversal := nat.NewTraversal(t.localID, target, t.localIP)

	localPort, err := t.puncher.Open()
	if err != nil {
		return nil, err
	}

	traversal.LocalPort = astral.Uint16(localPort)
	traversal.Session = t.puncher.Session()

	if err := ch.Send(traversal.OfferSignal()); err != nil {
		return nil, err
	}

	err = ch.Switch(
		traversal.ExpectSignal(nat.PunchSignalTypeAnswer, traversal.OnAnswer),
		channel.PassErrors,
		channel.WithContext(ctx),
	)
	if err != nil {
		return nil, err
	}

	if err := ch.Send(traversal.ReadySignal()); err != nil {
		return nil, err
	}

	err = ch.Switch(
		traversal.ExpectSignal(nat.PunchSignalTypeGo, nil),
		channel.PassErrors,
		channel.WithContext(ctx),
	)
	if err != nil {
		return nil, err
	}

	result, err := t.puncher.HolePunch(ctx, traversal.PeerIP, int(traversal.PeerPort))
	if err != nil {
		return nil, err
	}

	traversal.SetPunchResult(result)
	traversal.Pair.Nonce = astral.NewNonce()

	if err := ch.Send(traversal.ResultSignal()); err != nil {
		return nil, err
	}

	err = ch.Switch(
		traversal.ExpectSignal(nat.PunchSignalTypeResult, traversal.OnResult),
		channel.PassErrors,
		channel.WithContext(ctx),
	)
	if err != nil {
		return nil, err
	}

	return &traversal.Pair, nil
}
