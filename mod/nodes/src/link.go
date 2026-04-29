package nodes

import (
	"errors"
	"io"
	"math/rand"
	"sync/atomic"
	"time"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/mod/exonet"
	"github.com/cryptopunkscc/astrald/mod/nodes"
	"github.com/cryptopunkscc/astrald/sig"
)

type Ping struct {
	sentAt time.Time
	pong   chan struct{}
}

var _ nodes.Link = &Link{}
var _ nodes.QualityLink = &Link{}
var _ nodes.NetworkLink = &Link{}
var _ astral.Router = &Link{}

type Link struct {
	*channel.Channel
	id             astral.Nonce
	createdAt      time.Time
	localIdentity  *astral.Identity
	remoteIdentity *astral.Identity
	outbound       bool
	localEp        exonet.Endpoint
	remoteEp       exonet.Endpoint

	checks      atomic.Int32
	wakeCh      chan struct{}
	pingTimeout time.Duration
	pings       sig.Map[astral.Nonce, *Ping]
	pressure    LinkPressureDetector

	err  sig.Value[error]
	done chan struct{}

	Mux *Mux
}

func (s *Link) Done() <-chan struct{} { return s.done }

func (s *Link) Err() error { return s.err.Get() }

func (s *Link) LocalIdentity() *astral.Identity { return s.localIdentity }

func (s *Link) RemoteIdentity() *astral.Identity { return s.remoteIdentity }

func (s *Link) CloseWithError(err error) error {
	if err == nil {
		err = errors.New("link closed")
	}
	s.err.Swap(nil, err)
	_ = s.Channel.Close()
	return nil
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
	select {
	case s.wakeCh <- struct{}{}:
	default:
	}
}

func (s *Link) check() {
	if s.checks.Swap(2) != 0 {
		return
	}
	select {
	case s.wakeCh <- struct{}{}:
	default:
	}
}

func (s *Link) ping() (time.Duration, error) {
	if s.Mux == nil {
		return -1, nodes.ErrNotSupported
	}

	nonce := astral.NewNonce()
	p, ok := s.pings.Set(nonce, &Ping{
		sentAt: time.Now(),
		pong:   make(chan struct{}),
	})
	if !ok {
		return -1, errors.New("duplicate nonce")
	}
	defer s.pings.Delete(nonce)

	if err := s.Mux.ping(nonce); err != nil {
		return -1, err
	}
	p.sentAt = time.Now()

	select {
	case <-p.pong:
		rtt := time.Since(p.sentAt)
		if s.pressure != nil {
			s.pressure.OnRTT(rtt, time.Now())
		}
		return rtt, nil
	case <-time.After(s.pingTimeout):
		return -1, errors.New("ping timeout")
	case <-s.done:
		return -1, s.Err()
	}
}

func (s *Link) pong(nonce astral.Nonce) (time.Duration, error) {
	p, ok := s.pings.Delete(nonce)
	if !ok {
		return -1, errors.New("invalid nonce")
	}
	d := time.Since(p.sentAt)
	close(p.pong)
	return d, nil
}

func (s *Link) pingLoop() {
	select {
	case <-s.done:
		return
	case <-time.After(time.Duration(rand.Int63n(int64(pingJitter)))):
	}

	for {
		if s.checks.Load() > 0 {
			select {
			case <-s.done:
				return
			case <-time.After(activeInterval):
			}
		} else {
			select {
			case <-s.done:
				return
			case <-s.wakeCh:
			}
		}

		if s.Err() != nil {
			return
		}

		if _, err := s.ping(); err != nil {
			s.CloseWithError(err)
			return
		}

		if s.checks.Load() > 0 {
			s.checks.Add(-1)
		}
	}
}

func (s *Link) IsHighPressure() bool { return s.pressure != nil && s.pressure.IsHigh() }

func (s *Link) Throughput() uint64 { return s.Mux.throughput.Load() }

func (s *Link) RouteQuery(ctx *astral.Context, q *astral.InFlightQuery, w io.WriteCloser) (io.WriteCloser, error) {
	return s.Mux.RouteQuery(ctx, q, w)
}
