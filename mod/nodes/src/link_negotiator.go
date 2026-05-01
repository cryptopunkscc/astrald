package nodes

import (
	"context"
	"fmt"
	"slices"
	"time"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/mod/exonet"
	"github.com/cryptopunkscc/astrald/mod/nodes/src/noise"
	"github.com/decred/dcrd/dcrec/secp256k1/v4"
)

type muxLinkNegotiator struct {
	mod        *Module
	privateKey *secp256k1.PrivateKey
	features   []string
}

func (n *muxLinkNegotiator) OutboundHandshake(ctx context.Context, conn exonet.Conn, remoteID *astral.Identity) (*Link, error) {
	outbound, localEp, remoteEp := conn.Outbound(), conn.LocalEndpoint(), conn.RemoteEndpoint()

	aconn, err := noise.HandshakeOutbound(ctx, conn, remoteID.PublicKey(), n.privateKey)
	if err != nil {
		return nil, fmt.Errorf("outbound handshake: %w", err)
	}

	ch := channel.New(aconn)
	nonce, err := n.negotiateOutbound(ch)
	if err != nil {
		return nil, err
	}

	remoteIdentity := aconn.RemoteIdentity()
	localIdentity := aconn.LocalIdentity()

	return n.newLink(ch, localIdentity, remoteIdentity, nonce, outbound, localEp, remoteEp), nil
}

func (n *muxLinkNegotiator) InboundHandshake(ctx context.Context, conn exonet.Conn) (*Link, error) {
	outbound, localEp, remoteEp := conn.Outbound(), conn.LocalEndpoint(), conn.RemoteEndpoint()

	aconn, err := noise.HandshakeInbound(ctx, conn, n.privateKey)
	if err != nil {
		return nil, fmt.Errorf("inbound handshake: %w", err)
	}

	ch := channel.New(aconn)
	nonce, err := n.negotiateInbound(ch)
	if err != nil {
		return nil, err
	}

	remoteIdentity := aconn.RemoteIdentity()
	localIdentity := aconn.LocalIdentity()

	return n.newLink(ch, localIdentity, remoteIdentity, nonce, outbound, localEp, remoteEp), nil
}

func (n *muxLinkNegotiator) negotiateOutbound(ch *channel.Channel) (astral.Nonce, error) {
	var features []*astral.String8
	err := ch.Switch(
		channel.Collect[*astral.String8](&features),
		channel.BreakOnEOS,
	)
	if err != nil {
		return 0, fmt.Errorf("read features: %w", err)
	}

	var selected string
	for _, f := range n.features {
		if slices.ContainsFunc(features, func(pf *astral.String8) bool { return string(*pf) == f }) {
			selected = f
			break
		}
	}
	if selected == "" {
		return 0, fmt.Errorf("no supported link types found")
	}

	s := astral.String8(selected)
	if err = ch.Send(&s); err != nil {
		return 0, fmt.Errorf("write: %w", err)
	}

	var errCode *astral.Int8
	if err = ch.Switch(channel.Expect(&errCode)); err != nil {
		return 0, fmt.Errorf("read: %w", err)
	}
	if *errCode != 0 {
		return 0, fmt.Errorf("link feature negotation error")
	}

	var nonce *astral.Nonce
	if err = ch.Switch(channel.Expect(&nonce)); err != nil {
		return 0, fmt.Errorf("read outbound stream nonce: %w", err)
	}

	return *nonce, nil
}

func (n *muxLinkNegotiator) negotiateInbound(ch *channel.Channel) (astral.Nonce, error) {
	for _, feat := range n.features {
		s := astral.String8(feat)
		if err := ch.Send(&s); err != nil {
			return 0, err
		}
	}
	if err := ch.Send(&astral.EOS{}); err != nil {
		return 0, err
	}

	for {
		var selected *astral.String8
		if err := ch.Switch(channel.Expect(&selected)); err != nil {
			return 0, err
		}

		switch {
		case slices.Contains(n.features, string(*selected)):
			status := astral.Int8(0)
			if err := ch.Send(&status); err != nil {
				return 0, err
			}
			nonce := astral.NewNonce()
			if err := ch.Send(&nonce); err != nil {
				return 0, fmt.Errorf("failed to set inbound stream nonce: %w", err)
			}
			return nonce, nil
		default:
			status := astral.Int8(1)
			_ = ch.Send(&status)
			return 0, fmt.Errorf("unsupported feature requested: %s", *selected)
		}
	}
}

func (n *muxLinkNegotiator) newLink(ch *channel.Channel, localIdentity, remoteIdentity *astral.Identity, id astral.Nonce, outbound bool, localEp, remoteEp exonet.Endpoint) *Link {
	s := &Link{
		ch:             ch,
		id:             id,
		localIdentity:  localIdentity,
		remoteIdentity: remoteIdentity,
		createdAt:      time.Now(),
		outbound:       outbound,
		localEp:        localEp,
		remoteEp:       remoteEp,
		wakeCh:         make(chan struct{}, 1),
		pingTimeout:    defaultPingTimeout,
		done:           make(chan struct{}),
	}

	// todo: not sure if starting point should be here
	s.Mux = n.mod.newMux(ch)
	go s.Mux.Run(n.mod.ctx)

	return s
}
