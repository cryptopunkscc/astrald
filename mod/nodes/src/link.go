package nodes

import (
	"errors"
	"io"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/astral/log"
	"github.com/cryptopunkscc/astrald/mod/exonet"
	"github.com/cryptopunkscc/astrald/mod/nodes"
	"github.com/cryptopunkscc/astrald/sig"
)

var _ nodes.Link = &Link{}

type Link struct {
	id  astral.Nonce
	ch  *channel.Channel
	Mux *Mux

	//
	localIdentity  *astral.Identity
	remoteIdentity *astral.Identity
	outbound       bool
	localEp        exonet.Endpoint
	remoteEp       exonet.Endpoint
	//
	mu sync.Mutex
	// pressure
	pressure   nodes.LinkPressureDetector
	throughput atomic.Uint64
	// ping
	pingMu      sync.Mutex
	activePing  *Ping
	pingTimeout time.Duration
	log         *log.Logger
	logPings    bool
	//
	createdAt time.Time
	// lifecycle
	err    sig.Value[error]
	done   chan struct{}
	closed atomic.Bool
}

func (s *Link) ID() astral.Nonce { return s.id }

func (s *Link) Outbound() bool {
	return s.outbound
}

func (s *Link) SetRouter(r astral.Router) {
	s.Mux.SetRouter(r)
}

func (s *Link) Done() <-chan struct{} { return s.done }

func (s *Link) Err() error { return s.err.Get() }

func (s *Link) LocalIdentity() *astral.Identity { return s.localIdentity }

func (s *Link) RemoteIdentity() *astral.Identity { return s.remoteIdentity }

func (s *Link) CloseWithError(err error) error {
	if !s.closed.CompareAndSwap(false, true) {
		return nil
	}
	if err == nil {
		err = errors.New("link closed")
	}
	s.err.Swap(nil, err)
	_ = s.ch.Close()
	close(s.done)
	return nil
}

func (s *Link) Close() error {
	return s.CloseWithError(nil)
}

func (s *Link) Network() string {
	if s.remoteEp != nil {
		return s.remoteEp.Network()
	}
	if s.localEp != nil {
		return s.localEp.Network()
	}
	return "unknown"
}

func (s *Link) LocalEndpoint() exonet.Endpoint  { return s.localEp }
func (s *Link) RemoteEndpoint() exonet.Endpoint { return s.remoteEp }

func (s *Link) Wake() {
	s.pingMu.Lock()
	if s.activePing != nil {
		s.pingMu.Unlock()
		return
	}
	nonce := astral.NewNonce()
	p := &Ping{
		nonce:  nonce,
		sentAt: time.Now(),
		pong:   make(chan struct{}),
	}
	s.activePing = p
	s.pingMu.Unlock()

	go func() {
		defer func() {
			s.pingMu.Lock()
			s.activePing = nil
			s.pingMu.Unlock()
		}()

		if err := s.Mux.ping(nonce); err != nil {
			s.CloseWithError(err)
			return
		}
		p.sentAt = time.Now()

		select {
		case <-p.pong:
			rtt := time.Since(p.sentAt)
			s.OnRTT(rtt)
			if s.logPings {
				s.log.Logv(1, "ping with %v: %v", s.remoteIdentity, rtt)
			}
		case <-time.After(s.pingTimeout):
			s.CloseWithError(errors.New("ping timeout"))
		case <-s.done:
		}
	}()
}

func (s *Link) receivePong(nonce astral.Nonce) {
	s.pingMu.Lock()
	p := s.activePing
	s.pingMu.Unlock()
	if p == nil || p.nonce != nonce {
		return
	}
	close(p.pong)
}

func (s *Link) SetPressureDetector(d nodes.LinkPressureDetector) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.pressure = d
}

func (s *Link) AddThroughputBytes(n int) {
	s.throughput.Add(uint64(n))
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.pressure != nil {
		s.pressure.OnBytes(n, time.Now())
	}
}

func (s *Link) OnRTT(rtt time.Duration) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.pressure != nil {
		s.pressure.OnRTT(rtt, time.Now())
	}
}

func (s *Link) HasPressureDetector() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.pressure != nil
}

func (s *Link) IsHighPressure() bool {
	s.mu.Lock()
	defer s.mu.Unlock()
	return s.pressure != nil && s.pressure.IsHigh()
}

func (s *Link) Throughput() uint64 { return s.throughput.Load() }

func (s *Link) RouteQuery(ctx *astral.Context, q *astral.InFlightQuery, w io.WriteCloser) (io.WriteCloser, error) {
	return s.Mux.RouteQuery(ctx, q, w)
}

type Ping struct {
	nonce  astral.Nonce
	sentAt time.Time
	pong   chan struct{}
}
