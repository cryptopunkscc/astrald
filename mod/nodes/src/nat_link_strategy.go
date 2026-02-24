package nodes

import (
	"fmt"
	"sync"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/log"
	"github.com/cryptopunkscc/astrald/lib/astrald"
	"github.com/cryptopunkscc/astrald/mod/kcp"
	kcpclient "github.com/cryptopunkscc/astrald/mod/kcp/client"
	natclient "github.com/cryptopunkscc/astrald/mod/nat/client"
	"github.com/cryptopunkscc/astrald/mod/nodes"
)

type NatLinkStrategy struct {
	mod    *Module
	log    *log.Logger
	target *astral.Identity

	mu   sync.Mutex
	done chan struct{}
}

var _ nodes.LinkStrategy = &NatLinkStrategy{}

func (s *NatLinkStrategy) Name() string { return nodes.StrategyNAT }

func (s *NatLinkStrategy) Signal(ctx *astral.Context) {
	s.mu.Lock()
	if s.done != nil {
		s.mu.Unlock()
		return
	}
	s.done = make(chan struct{})
	s.mu.Unlock()

	go func() {
		defer s.signalDone()

		if err := s.attempt(ctx); err != nil {
			s.log.Logv(2, "%v %v", s.target, err)
		}
	}()
}

func (s *NatLinkStrategy) attempt(ctx *astral.Context) error {
	selfID := s.mod.node.Identity()
	ctx = ctx.IncludeZone(astral.ZoneNetwork)

	s.log.Log("%v starting traversal", s.target)

	natClient := natclient.New(selfID, astrald.Default())
	pair, err := natClient.Traverse(ctx, s.target)
	if err != nil {
		return fmt.Errorf("traverse: %w", err)
	}

	s.log.Log("%v traversal complete, locking pair %v", s.target, pair.Nonce)
	if err := natClient.PairTake(ctx, pair.Nonce, s.target); err != nil {
		return fmt.Errorf("pair take: %w", err)
	}

	s.log.Log("%v pair locked, setting up kcp", s.target)
	local, remote := pair.PeerA, pair.PeerB
	if pair.PeerB.Identity.IsEqual(selfID) {
		local, remote = pair.PeerB, pair.PeerA
	}

	peerEndpoint := kcp.Endpoint{
		IP:   remote.Endpoint.IP,
		Port: remote.Endpoint.Port,
	}

	localEndpoint := kcp.Endpoint{
		IP:   local.Endpoint.IP,
		Port: local.Endpoint.Port,
	}

	kcpClient := kcpclient.New(selfID, astrald.Default())

	// Set up the remote side: ephemeral listener + endpoint mapping
	err = kcpClient.WithTarget(s.target).CreateEphemeralListener(ctx, peerEndpoint.Port)
	if err != nil {
		return fmt.Errorf("remote create ephemeral listener: %w", err)
	}

	err = kcpClient.WithTarget(s.target).SetEndpointLocalPort(ctx, localEndpoint, peerEndpoint.Port, true)
	if err != nil {
		return fmt.Errorf("remote set endpoint local port: %w", err)
	}

	err = kcpClient.SetEndpointLocalPort(ctx, peerEndpoint, localEndpoint.Port, true)
	if err != nil {
		return fmt.Errorf("set endpoint local port: %w", err)
	}

	s.log.Log("%v dialing %v", s.target, peerEndpoint.Address())
	conn, err := s.mod.Exonet.Dial(ctx, &peerEndpoint)
	if err != nil {
		return fmt.Errorf("dial kcp: %w", err)
	}

	stream, err := s.mod.peers.EstablishOutboundLink(ctx, s.target, conn)
	if err != nil {
		conn.Close()
		return fmt.Errorf("establish link: %w", err)
	}

	target := s.target
	stream.OnClose(func() {
		cleanupCtx := s.mod.ctx.
			IncludeZone(astral.ZoneNetwork)

		if err := kcpClient.RemoveEndpointLocalPort(cleanupCtx, peerEndpoint); err != nil {
			s.log.Logv(2, "cleanup local socket mapping: %v", err)
		}
		if err := kcpClient.WithTarget(target).CloseEphemeralListener(cleanupCtx, peerEndpoint.Port); err != nil {
			s.log.Logv(2, "cleanup remote ephemeral listener: %v", err)
		}
		if err := kcpClient.WithTarget(target).RemoveEndpointLocalPort(cleanupCtx, localEndpoint); err != nil {
			s.log.Logv(2, "cleanup remote socket mapping: %v", err)
		}
	})

	s.log.Log("%v linked via %v", s.target, peerEndpoint.Address())
	name := s.Name()
	if !s.mod.linkPool.notifyStreamWatchers(stream, &name) {
		stream.CloseWithError(nodes.ErrExcessStream)
	}

	return nil
}

func (s *NatLinkStrategy) signalDone() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.done != nil {
		close(s.done)
		s.done = nil
	}
}

func (s *NatLinkStrategy) Done() <-chan struct{} {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.done == nil {
		ch := make(chan struct{})
		close(ch)
		return ch
	}
	return s.done
}

// factory

type NatLinkStrategyFactory struct {
	mod *Module
}

var _ nodes.StrategyFactory = &NatLinkStrategyFactory{}

func (f *NatLinkStrategyFactory) Build(target *astral.Identity) nodes.LinkStrategy {
	return &NatLinkStrategy{
		mod:    f.mod,
		log:    f.mod.log.AppendTag("nat"),
		target: target,
	}
}
