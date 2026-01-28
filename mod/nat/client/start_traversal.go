package nat

import (
	"time"

	"github.com/cryptopunkscc/astrald/astral"
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
	Punch         func(remoteIP ip.IP, remotePort int) (ip.IP, int, error)
}

func (client *Client) NewTraversalClient(localIP ip.IP, session []byte) (*TraversalClient, error) {
	return &TraversalClient{
		LocalIdentity: client.astral.GuestID(),
		LocalIP:       localIP,
		Session:       session,
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

	tr := nat.NewTraversal(ch, ctx.Identity(), target, localIP, nil, nil)

	if err := ch.Send(&nat.PunchSignal{
		Signal:  nat.PunchSignalTypeOffer,
		Session: session,
		IP:      localIP,
		Port:    astral.Uint16(localPort),
	}); err != nil {
		return nil, err
	}

	signal, err := tr.Expect(nat.PunchSignalTypeAnswer)
	if err != nil {
		return nil, err
	}

	err = ch.Send(&nat.PunchSignal{
		Signal:  nat.PunchSignalTypeReady,
		Session: signal.Session,
	})
	if err != nil {
		return nil, err
	}

	if _, err := tr.Expect(nat.PunchSignalTypeGo); err != nil {
		return nil, err
	}

	remoteIP, remotePort, err := t.Punch(signal.IP, int(signal.Port))
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

	resSig, err := tr.Expect(nat.PunchSignalTypeResult)
	if err != nil {
		return nil, err
	}
	pair.PeerA.Endpoint = nat.UDPEndpoint{IP: resSig.IP, Port: resSig.Port}
	pair.Nonce = resSig.PairNonce

	return pair, nil
}
