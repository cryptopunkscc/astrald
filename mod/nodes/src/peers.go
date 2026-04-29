package nodes

import (
	"context"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/exonet"
	"github.com/cryptopunkscc/astrald/mod/nodes"
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
	streamsWithSameIdentity := mod.linkPool.Links().Select(func(v *Link) bool {
		return v.RemoteIdentity().IsEqual(s.RemoteIdentity())
	})

	if !s.outbound {
		mod.linkPool.notifyLinkWatchers(s, nil)
	}

	mod.Events.Emit(&nodes.StreamCreatedEvent{
		RemoteIdentity: s.RemoteIdentity(),
		StreamId:       s.id,
		StreamCount:    len(streamsWithSameIdentity),
	})

	go func() {
		s.Mux.Run(mod.ctx)

		tl.Close()

		streamsWithSameIdentity := mod.linkPool.Links().Select(func(v *Link) bool {
			return v.RemoteIdentity().IsEqual(s.RemoteIdentity())
		})

		mod.Events.Emit(&nodes.StreamClosedEvent{
			RemoteIdentity: s.RemoteIdentity(),
			Forced:         false,
			StreamCount:    astral.Int8(len(streamsWithSameIdentity)),
		})

		mod.log.Info("closed %v-stream with %v (%v): %v", dir, s.RemoteIdentity(), netName, s.Err())
	}()

	go mod.reflectLink(s)

	return
}

// isLinked returns true if there is at least one stream to the given identity.
func (mod *Peers) isLinked(remoteID *astral.Identity) bool {
	for _, s := range mod.linkPool.Links().Clone() {
		if s.RemoteIdentity().IsEqual(remoteID) {
			return true
		}
	}
	return false
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
