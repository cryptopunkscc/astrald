package nat

import (
	"time"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/query"
	"github.com/cryptopunkscc/astrald/mod/ip"
	"github.com/cryptopunkscc/astrald/mod/nat"
)

type TraversalClient struct {
	client        *Client
	LocalIdentity *astral.Identity
	LocalIP       ip.IP
	LocalPort     uint16
	Session       []byte
	Puncher       nat.Puncher
}

func (client *Client) NewTraversalClient(localIP ip.IP, session []byte, puncher nat.Puncher) (*TraversalClient, error) {
	return &TraversalClient{
		LocalIdentity: client.astral.GuestID(),
		LocalIP:       localIP,
		Session:       session,
		puncher:       puncher,
	}, nil
}

func (t *TraversalClient) StartTraversal(ctx *astral.Context, target *astral.Identity) (*nat.TraversedPortPair, error) {
	ch, err := t.client.queryCh(ctx, nat.MethodStartNatTraversal, query.Args{
		"target": target.String(),
	})
	if err != nil {
		return nil, err
	}
	defer ch.Close()

	session := t.Session
	localIP := t.LocalIP
	localPort := t.LocalPort

	tr := nat.NewTraversal(ch, ctx.Identity(), target, localIP, t.Puncher)

	err = tr.SendOffer(localIP)
	if err != nil {
		return nil, err
	}

	err = ch.Switch(
		tr.ExpectPunchSignal(nat.PunchSignalTypeOffer, tr.OnAnswer),
		channel.PassErrors,
		channel.WithContext(ctx),
	)
	if err != nil {
		return nil, err
	}

	err = tr.SendReady()
	if err != nil {
		return nil, err
	}

	err = ch.Switch(
		tr.ExpectPunchSignal(nat.PunchSignalTypeOffer, nil),
		channel.PassErrors,
		channel.WithContext(ctx),
	)
	if err != nil {
		return nil, err
	}

	result, err := t.Puncher.HolePunch(ctx, tr.PeerIP, int(tr.PeerPort))
	if err != nil {
		return nil, err
	}

	pair := &nat.TraversedPortPair{
		Nonce:     astral.NewNonce(),
		CreatedAt: astral.Time(time.Now()),
		PeerA:     nat.PeerEndpoint{Identity: ctx.Identity()},
		PeerB:     nat.PeerEndpoint{Identity: target, Endpoint: nat.UDPEndpoint{IP: remoteIP, Port: astral.Uint16(remotePort)}},
	}

	if err := tr.SendResult(pair.PeerB.Endpoint.IP, uint16(pair.PeerB.Endpoint.Port), pair.Nonce); err != nil {
		return nil, err
	}

	err = ch.Switch(
		tr.ExpectPunchSignal(nat.PunchSignalTypeResult, tr.OnResult),
		channel.PassErrors,
		channel.WithContext(ctx),
	)
	if err != nil {
		return nil, err
	}

	return pair, nil
}
