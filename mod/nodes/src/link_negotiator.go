package nodes

import (
	"fmt"
	"slices"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/astral/log"
	"github.com/cryptopunkscc/astrald/mod/nodes/src/noise"
)

type linkNegotiator struct {
	log      *log.Logger
	features []string
}

func newLinkNegotiator(log *log.Logger, features []string) *linkNegotiator {
	return &linkNegotiator{log: log, features: features}
}

func (n *linkNegotiator) NegotiateOutbound(aconn *noise.Conn) (*Link, error) {
	ch := channel.New(aconn)

	var features []*astral.String8
	err := ch.Switch(
		channel.Collect[*astral.String8](&features),
		channel.BreakOnEOS,
	)
	if err != nil {
		return nil, fmt.Errorf("read features: %w", err)
	}

	var selected string
	for _, f := range n.features {
		if slices.ContainsFunc(features, func(pf *astral.String8) bool { return string(*pf) == f }) {
			selected = f
			break
		}
	}
	if selected == "" {
		return nil, fmt.Errorf("no supported link types found")
	}

	s := astral.String8(selected)
	if err = ch.Send(&s); err != nil {
		return nil, fmt.Errorf("write: %w", err)
	}

	var errCode *astral.Int8
	if err = ch.Switch(channel.Expect(&errCode)); err != nil {
		return nil, fmt.Errorf("read: %w", err)
	}
	if *errCode != 0 {
		return nil, fmt.Errorf("link feature negotation error")
	}

	var nonce *astral.Nonce
	if err = ch.Switch(channel.Expect(&nonce)); err != nil {
		return nil, fmt.Errorf("read outbound stream nonce: %w", err)
	}

	return newLink(aconn, *nonce, true), nil
}

func (n *linkNegotiator) NegotiateInbound(aconn *noise.Conn) (*Link, error) {
	ch := channel.New(aconn)

	for _, feat := range n.features {
		s := astral.String8(feat)
		if err := ch.Send(&s); err != nil {
			return nil, err
		}
	}
	if err := ch.Send(&astral.EOS{}); err != nil {
		return nil, err
	}

	for {
		var selected *astral.String8
		if err := ch.Switch(channel.Expect(&selected)); err != nil {
			return nil, err
		}

		switch {
		case slices.Contains(n.features, string(*selected)):
			status := astral.Int8(0)
			if err := ch.Send(&status); err != nil {
				return nil, err
			}
			nonce := astral.NewNonce()
			if err := ch.Send(&nonce); err != nil {
				return nil, fmt.Errorf("failed to set inbound stream nonce: %w", err)
			}
			return newLink(aconn, nonce, false), nil
		default:
			status := astral.Int8(1)
			_ = ch.Send(&status)
			return nil, fmt.Errorf("remote party (%s from %s) requested an invalid feature: %s",
				aconn.RemoteIdentity(),
				aconn.RemoteEndpoint(),
				*selected,
			)
		}
	}
}
