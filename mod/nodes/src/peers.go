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

func (mod *Peers) peers() (peers []*astral.Identity) {
	var seen = map[string]struct{}{}
	for _, s := range mod.linkPool.Links().Clone() {
		key := s.RemoteIdentity().String()
		if _, found := seen[key]; found {
			continue
		}
		seen[key] = struct{}{}
		peers = append(peers, s.RemoteIdentity())
	}
	return
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

	tl, err := mod.linkPool.AddLink(s)
	if err != nil {
		return
	}

	mod.log.Infov(1, "added %v-stream with %v (%v)", dir, s.RemoteIdentity(), netName)

	go func() {
		s.Mux.Run(mod.ctx)
		tl.Close()
		mod.log.Info("closed %v-stream with %v (%v): %v", dir, s.RemoteIdentity(), netName, s.Err())
	}()

	go mod.reflectLink(s)

	return
}

func (mod *Peers) EstablishOutboundLink(ctx context.Context, remoteID *astral.Identity, conn exonet.Conn) (_ *Link, err error) {
	defer func() {
		if err != nil {
			conn.Close()
		}
	}()

	negotiator, err := mod.GetLinkNegotiator()
	if err != nil {
		return nil, err
	}

	link, err := negotiator.OutboundHandshake(ctx, conn, remoteID)
	if err != nil {
		return nil, err
	}

	err = mod.addLink(link)
	return link, err
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

	stream, err := negotiator.InboundHandshake(ctx, conn)
	if err != nil {
		return err
	}

	return mod.addLink(stream)
}
