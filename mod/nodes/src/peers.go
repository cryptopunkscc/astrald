package nodes

import (
	"context"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/exonet"
)

type Peers struct {
	*Module
}

func NewPeers(m *Module) *Peers {
	return &Peers{Module: m}
}

func (mod *Peers) addLink(s *Link) (err error) {
	var dir = "in"
	var netName = "unknown network"

	if s.outbound {
		dir = "out"
	}

	switch {
	case s.LocalEndpoint() != nil:
		netName = s.LocalEndpoint().Network()
	case s.RemoteEndpoint() != nil:
		netName = s.RemoteEndpoint().Network()
	}

	_, err = mod.linkPool.AddLink(s)
	if err != nil {
		return
	}

	mod.log.Infov(1, "added %v-link with %v (%v)", dir, s.RemoteIdentity(), netName)

	go func() {
		<-s.Done()
		mod.log.Info("closed %v-link with %v (%v): %v", dir, s.RemoteIdentity(), netName, s.Err())
	}()

	go mod.reflectLink(s)

	return
}

func (mod *Peers) EstablishOutboundLink(ctx context.Context, remoteID *astral.Identity, conn exonet.Conn) (link *Link, err error) {
	defer func() {
		if err != nil {
			conn.Close()
		}
	}()

	negotiator, err := mod.GetLinkNegotiator()
	if err != nil {
		return nil, err
	}

	negotiatedLink, err := negotiator.OutboundHandshake(ctx, conn, remoteID)
	if err != nil {
		return nil, err
	}

	link, err = mod.linkPool.AddLink(negotiatedLink)
	if err != nil {
		return nil, err
	}

	return link, nil
}

func (mod *Peers) EstablishInboundLink(ctx context.Context, conn exonet.Conn) (err error) {
	defer func() {
		if err != nil {
			conn.Close()
		}
	}()

	negotiator, err := mod.GetLinkNegotiator()
	if err != nil {
		return err
	}

	link, err := negotiator.InboundHandshake(ctx, conn)
	if err != nil {
		return err
	}

	_, err = mod.linkPool.AddLink(link)
	if err != nil {
		return err
	}

	return nil
}
