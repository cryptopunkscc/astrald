package nodes

import (
	"sync"
	"time"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/exonet"
	"github.com/cryptopunkscc/astrald/mod/nodes"
)

// PersistentLinkStrategy tries to connect with retries. After signalTimeout it
// signals Done but continues running in background on module context for up to
// backgroundTimeout more. Useful for slow networks like Tor.
type PersistentLinkStrategy struct {
	mod               *Module
	network           string
	target            *astral.Identity
	quickRetries      int // immediate retries (no delay)
	retries           int // normal retries (with delay)
	retryDelay        time.Duration
	signalTimeout     time.Duration
	backgroundTimeout time.Duration

	mu         sync.Mutex
	activeDone chan struct{}
}

var _ nodes.LinkStrategy = &PersistentLinkStrategy{}

func (s *PersistentLinkStrategy) Signal(ctx *astral.Context) {
	s.mu.Lock()
	if s.activeDone != nil {
		s.mu.Unlock()
		s.mod.log.Logv(1, "[%v] %v: signal ignored, already running", s.network, s.target)
		return
	}
	s.activeDone = make(chan struct{})
	s.mu.Unlock()

	s.mod.log.Log("[%v] %v: starting link strategy", s.network, s.target)
	go s.run(ctx)
}

func (s *PersistentLinkStrategy) run(ctx *astral.Context) {
	signalTimer := time.NewTimer(s.signalTimeout)
	defer signalTimer.Stop()

	resultCh := make(chan *Stream, 1)

	// Start worker with caller's context initially
	workerCtx, workerCancel := ctx.WithCancel()
	go func() {
		defer close(resultCh)
		if stream := s.tryWithRetry(workerCtx); stream != nil {
			resultCh <- stream
		}
	}()

	select {
	case stream := <-resultCh:
		workerCancel()
		s.signalDone()
		if stream != nil {
			s.mod.log.Log("[%v] %v: connected before timeout", s.network, s.target)
		} else {
			s.mod.log.Log("[%v] %v: all attempts failed before timeout", s.network, s.target)
		}
		s.deliverStream(stream)
		return

	case <-ctx.Done():
		workerCancel()
		s.signalDone()
		s.mod.log.Log("[%v] %v: cancelled by caller", s.network, s.target)
		return

	case <-signalTimer.C:
		workerCancel()
		s.signalDone()
		s.mod.log.Log("[%v] %v: signal timeout (%v), continuing in background for %v",
			s.network, s.target, s.signalTimeout, s.backgroundTimeout)
	}

	// Continue in background on module context
	bgCtx, bgCancel := s.mod.ctx.WithTimeout(s.backgroundTimeout)
	defer bgCancel()

	bgResultCh := make(chan *Stream, 1)
	go func() {
		defer close(bgResultCh)
		if stream := s.tryWithRetry(bgCtx); stream != nil {
			bgResultCh <- stream
		}
	}()

	select {
	case stream := <-bgResultCh:
		if stream != nil {
			s.mod.log.Log("[%v] %v: connected in background mode", s.network, s.target)
		} else {
			s.mod.log.Logv(1, "[%v] %v: background attempts exhausted", s.network, s.target)
		}
		s.deliverStream(stream)
	case <-bgCtx.Done():
		s.mod.log.Logv(1, "[%v] %v: background timeout expired", s.network, s.target)
	}
}

func (s *PersistentLinkStrategy) signalDone() {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.activeDone != nil {
		close(s.activeDone)
		s.activeDone = nil
	}
}

func (s *PersistentLinkStrategy) deliverStream(stream *Stream) {
	if stream == nil {
		return
	}
	if !s.mod.linkPool.notifyStreamWatchers(stream) {
		s.mod.log.Logv(1, "[%v] %v: stream not consumed, closing as excess",
			s.network, s.target)
		stream.CloseWithError(nodes.ErrExcessStream)
	} else {
		s.mod.log.Logv(1, "[%v] %v: stream delivered to watcher", s.network, s.target)
	}
}

func (s *PersistentLinkStrategy) tryWithRetry(ctx *astral.Context) *Stream {
	attempt := 0
	totalAttempts := s.quickRetries + s.retries

	// Quick retries first (no delay) - handles circuit building failures
	for i := 0; i < s.quickRetries; i++ {
		attempt++
		if ctx.Err() != nil {
			return nil
		}
		if stream := s.tryOnce(ctx, attempt, totalAttempts); stream != nil {
			return stream
		}
	}

	// Normal retries with delay
	for i := 0; i < s.retries; i++ {
		attempt++
		if ctx.Err() != nil {
			return nil
		}
		if stream := s.tryOnce(ctx, attempt, totalAttempts); stream != nil {
			return stream
		}

		// Don't wait after last attempt
		if i < s.retries-1 {
			s.mod.log.Logv(1, "[%v] %v: retrying in %v", s.network, s.target, s.retryDelay)
			select {
			case <-time.After(s.retryDelay):
			case <-ctx.Done():
				return nil
			}
		}
	}
	return nil
}

func (s *PersistentLinkStrategy) tryOnce(ctx *astral.Context, attempt, total int) *Stream {
	resolved, err := s.mod.ResolveEndpoints(ctx, s.target)
	if err != nil {
		s.mod.log.Logv(1, "[%v] %v: attempt %v/%v failed to resolve endpoints: %v",
			s.network, s.target, attempt, total, err)
		return nil
	}

	endpointCount := 0
	for endpoint := range resolved {
		if ctx.Err() != nil {
			return nil
		}

		if endpoint.Network() != s.network {
			continue
		}

		endpointCount++
		if stream := s.tryEndpoint(ctx, endpoint, attempt, total); stream != nil {
			return stream
		}
	}

	if endpointCount == 0 {
		s.mod.log.Logv(1, "[%v] %v: attempt %v/%v found no endpoints",
			s.network, s.target, attempt, total)
	}

	return nil
}

func (s *PersistentLinkStrategy) tryEndpoint(ctx *astral.Context, endpoint exonet.Endpoint, attempt, total int) *Stream {
	s.mod.log.Logv(1, "[%v] %v: attempt %v/%v dialing %v",
		s.network, s.target, attempt, total, endpoint)

	conn, err := s.mod.Exonet.Dial(ctx, endpoint)
	if err != nil {
		s.mod.log.Logv(1, "[%v] %v: attempt %v/%v dial failed: %v",
			s.network, s.target, attempt, total, err)
		return nil
	}

	s.mod.log.Logv(1, "[%v] %v: attempt %v/%v connected, establishing link",
		s.network, s.target, attempt, total)

	stream, err := s.mod.peers.EstablishOutboundLink(ctx, s.target, conn)
	if err != nil {
		s.mod.log.Logv(1, "[%v] %v: attempt %v/%v link failed: %v",
			s.network, s.target, attempt, total, err)
		conn.Close()
		return nil
	}

	s.mod.log.Log("[%v] %v: link established via %v", s.network, s.target, endpoint)
	return stream
}

func (s *PersistentLinkStrategy) Done() <-chan struct{} {
	s.mu.Lock()
	defer s.mu.Unlock()

	if s.activeDone == nil {
		ch := make(chan struct{})
		close(ch)
		return ch
	}
	return s.activeDone
}

// Factory

type PersistentLinkStrategyConfig struct {
	QuickRetries      int // immediate retries (no delay) for circuit building
	Retries           int // normal retries with delay
	RetryDelay        time.Duration
	SignalTimeout     time.Duration
	BackgroundTimeout time.Duration
}

type PersistentLinkStrategyFactory struct {
	mod     *Module
	network string
	config  PersistentLinkStrategyConfig
}

var _ nodes.StrategyFactory = &PersistentLinkStrategyFactory{}

func (f *PersistentLinkStrategyFactory) Build(target *astral.Identity) nodes.LinkStrategy {
	return &PersistentLinkStrategy{
		mod:               f.mod,
		network:           f.network,
		target:            target,
		quickRetries:      f.config.QuickRetries,
		retries:           f.config.Retries,
		retryDelay:        f.config.RetryDelay,
		signalTimeout:     f.config.SignalTimeout,
		backgroundTimeout: f.config.BackgroundTimeout,
	}
}
