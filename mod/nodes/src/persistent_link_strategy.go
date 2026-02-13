package nodes

import (
	"sync"
	"time"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/log"
	"github.com/cryptopunkscc/astrald/mod/exonet"
	"github.com/cryptopunkscc/astrald/mod/nodes"
	"github.com/cryptopunkscc/astrald/sig"
)

// PersistentLinkStrategy tries to connect with retries. After signalTimeout it
// signals Done but continues running in background on module context for up to
// backgroundTimeout more. Useful for slow networks like Tor.
type PersistentLinkStrategy struct {
	mod               *Module
	log               *log.Logger
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
		return
	}
	s.activeDone = make(chan struct{})
	s.mu.Unlock()

	go s.run(ctx)
}

func (s *PersistentLinkStrategy) run(ctx *astral.Context) {
	// Resolve endpoints once upfront
	resolved, err := s.mod.ResolveEndpoints(ctx, s.target)
	if err != nil {
		s.log.Logv(2, "%v resolve failed: %v", s.target, err)
		s.signalDone()
		return
	}
	filtered := sig.FilterChan(resolved, func(e exonet.Endpoint) bool {
		return e.Network() == s.network
	})
	endpoints := sig.ChanToArray(filtered)
	if len(endpoints) == 0 {
		s.log.Logv(2, "%v no endpoints found to", s.target)
		s.signalDone()
		return
	}

	signalTimer := time.NewTimer(s.signalTimeout)
	defer signalTimer.Stop()

	resultCh := make(chan *Stream, 1)

	// Start worker with caller's context initially
	workerCtx, workerCancel := ctx.WithCancel()
	go func() {
		defer close(resultCh)
		if stream := s.tryWithRetry(workerCtx, endpoints); stream != nil {
			resultCh <- stream
		}
	}()

	select {
	case stream := <-resultCh:
		workerCancel()
		s.signalDone()
		s.deliverStream(stream)
		return
	case <-ctx.Done():
		workerCancel()
		s.signalDone()
		return

	case <-signalTimer.C:
		workerCancel()
		s.signalDone()
	}

	// Continue in background on module context
	bgCtx, bgCancel := s.mod.ctx.WithTimeout(s.backgroundTimeout)
	defer bgCancel()

	bgResultCh := make(chan *Stream, 1)
	go func() {
		defer close(bgResultCh)
		if stream := s.tryWithRetry(bgCtx, endpoints); stream != nil {
			bgResultCh <- stream
		}
	}()

	select {
	case stream := <-bgResultCh:
		s.deliverStream(stream)
	case <-bgCtx.Done():
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
		stream.CloseWithError(nodes.ErrExcessStream)
	}
}

func (s *PersistentLinkStrategy) tryWithRetry(ctx *astral.Context, endpoints []exonet.Endpoint) *Stream {
	// Quick retries (no delay)
	for i := 0; i < s.quickRetries; i++ {
		for _, ep := range endpoints {
			if ctx.Err() != nil {
				return nil
			}
			if stream := s.tryEndpoint(ctx, ep); stream != nil {
				return stream
			}
		}
		s.log.Logv(2, "%v quick retry %d/%d", s.target, i+1, s.quickRetries)
	}

	// Normal retries with delay
	for i := 0; i < s.retries; i++ {
		for _, ep := range endpoints {
			if ctx.Err() != nil {
				return nil
			}
			if stream := s.tryEndpoint(ctx, ep); stream != nil {
				return stream
			}
		}
		if i < s.retries-1 {
			s.log.Logv(2, "%v retry %d/%d in %v", s.target, i+1, s.retries, s.retryDelay)
			select {
			case <-time.After(s.retryDelay):
			case <-ctx.Done():
				return nil
			}
		}
	}
	return nil
}

func (s *PersistentLinkStrategy) tryEndpoint(ctx *astral.Context, endpoint exonet.Endpoint) *Stream {
	conn, err := s.mod.Exonet.Dial(ctx, endpoint)
	if err != nil {
		s.log.Logv(2, "%v dial %v: %v", s.target, endpoint, err)
		return nil
	}

	stream, err := s.mod.peers.EstablishOutboundLink(ctx, s.target, conn)
	if err != nil {
		s.log.Logv(2, "%v link %v: %v", s.target, endpoint, err)
		conn.Close()
		return nil
	}

	s.log.Log("%v linked via %v", s.target, endpoint)
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
		log:               f.mod.log.AppendTag(log.Tag(f.network)),
		network:           f.network,
		target:            target,
		quickRetries:      f.config.QuickRetries,
		retries:           f.config.Retries,
		retryDelay:        f.config.RetryDelay,
		signalTimeout:     f.config.SignalTimeout,
		backgroundTimeout: f.config.BackgroundTimeout,
	}
}
