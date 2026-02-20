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
	ac := astrald.Default()

	s.log.Log("%v starting traversal", s.target)

	nc := natclient.New(selfID, ac)
	pair, err := nc.Traverse(ctx, s.target)
	if err != nil {
		return fmt.Errorf("traverse: %w", err)
	}

	s.log.Log("%v traversal complete, locking pair %v", s.target, pair.Nonce)

	natClient := natclient.New(s.target, ac)
	pairCtx, cancelPair := ctx.WithCancel()
	defer cancelPair()

	// note: rethink
	errs := make(chan error, 2)
	go func() { errs <- nc.PairLock(pairCtx, pair.Nonce) }()
	go func() { errs <- natClient.PairTake(pairCtx, pair.Nonce) }()

	for range 2 {
		if err := <-errs; err != nil {
			cancelPair()
			return fmt.Errorf("pair lock/take: %w", err)
		}
	}

	s.log.Log("%v pair locked, setting up kcp", s.target)
	local, remote := pair.PeerA, pair.PeerB
	if pair.PeerB.Identity.IsEqual(selfID) {
		local, remote = pair.PeerB, pair.PeerA
	}

	localPort := local.Endpoint.Port
	peerEp := kcp.Endpoint{
		IP:   remote.Endpoint.IP,
		Port: remote.Endpoint.Port,
	}

	kcpClient := kcpclient.New(selfID, ac)

	if err := kcpClient.CreateEphemeralListener(ctx, localPort); err != nil {
		return fmt.Errorf("create ephemeral listener: %w", err)
	}

	// note: closing ephemeral listener do not close existing streams
	defer kcpClient.CloseEphemeralListener(ctx, localPort)

	if err := kcpClient.SetEndpointLocalPort(ctx, peerEp, localPort, true); err != nil {
		return fmt.Errorf("set endpoint local port: %w", err)
	}

	s.log.Log("%v dialing %v", s.target, peerEp.Address())
	conn, err := s.mod.Exonet.Dial(ctx, &peerEp)
	if err != nil {
		return fmt.Errorf("dial kcp: %w", err)
	}

	stream, err := s.mod.peers.EstablishOutboundLink(ctx, s.target, conn)
	if err != nil {
		conn.Close()
		return fmt.Errorf("establish link: %w", err)
	}

	s.log.Log("%v linked via %v", s.target, peerEp.Address())
	if !s.mod.linkPool.notifyStreamWatchers(stream) {
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
		log:    f.mod.log.AppendTag(log.Tag("nat")),
		target: target,
	}
}
