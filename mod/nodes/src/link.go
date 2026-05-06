package nodes

import (
	"errors"
	"io"
	"math/rand"
	"sync/atomic"
	"time"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/exonet"
	"github.com/cryptopunkscc/astrald/mod/nodes/frames"
	"github.com/cryptopunkscc/astrald/sig"
)

type Link struct {
	*frames.Stream
	mux         *Mux
	id          astral.Nonce
	createdAt   time.Time
	conn        astral.Conn
	pings       sig.Map[astral.Nonce, *Ping]
	checks      atomic.Int32
	throughput  atomic.Uint64
	outbound    bool
	pingTimeout time.Duration
	wakeCh      chan struct{}
	pressure    LinkPressureDetector
}

type Ping struct {
	sentAt time.Time
	pong   chan struct{}
}

func newLink(mod *Module, conn astral.Conn, id astral.Nonce, outbound bool) *Link {
	link := &Link{
		id:          id,
		conn:        conn,
		createdAt:   time.Now(),
		Stream:      frames.NewStream(conn),
		outbound:    outbound,
		pingTimeout: defaultPingTimeout,
		wakeCh:      make(chan struct{}, 1),
	}
	link.mux = newMux(mod, link)

	go link.pingLoop()

	return link
}

func (s *Link) LocalIdentity() *astral.Identity {
	return s.conn.LocalIdentity()
}

func (s *Link) RemoteIdentity() *astral.Identity {
	return s.conn.RemoteIdentity()
}

func (s *Link) CloseWithError(err error) error {
	if err != nil {
		return s.Stream.CloseWithError(err)
	}

	return s.Stream.CloseWithError(errors.New("link closed"))
}

func (s *Link) Close() error {
	return s.CloseWithError(nil)
}

func (s *Link) Network() string {
	if c, ok := s.conn.(exonet.Conn); ok {
		if e := c.RemoteEndpoint(); e != nil {
			return e.Network()
		}
		if e := c.LocalEndpoint(); e != nil {
			return e.Network()
		}
	}
	return "unknown"
}

func (s *Link) LocalEndpoint() exonet.Endpoint {
	if c, ok := s.conn.(exonet.Conn); ok {
		return c.LocalEndpoint()
	}
	return nil
}

func (s *Link) RemoteEndpoint() exonet.Endpoint {
	if c, ok := s.conn.(exonet.Conn); ok {
		return c.RemoteEndpoint()
	}

	return nil
}

func (s *Link) Outbound() bool {
	return s.outbound
}

func (s *Link) PressureHigh() bool {
	return s.pressure != nil && s.pressure.IsHigh()
}

func (s *Link) Throughput() uint64 {
	return s.throughput.Load()
}

func (s *Link) RouteQuery(ctx *astral.Context, q *astral.InFlightQuery, w io.WriteCloser) (io.WriteCloser, error) {
	return s.mux.RouteQuery(ctx, q, w)
}

func (s *Link) Write(frame frames.Frame) (err error) {
	if f, ok := frame.(*frames.Data); ok {
		s.throughput.Add(uint64(len(f.Payload)))
		if s.pressure != nil {
			s.pressure.OnBytes(len(f.Payload), time.Now())
		}
	}
	if _, ok := frame.(*frames.Ping); !ok {
		s.check()
	}
	return s.Stream.Write(frame)
}

func (s *Link) Ping() (time.Duration, error) {
	var nonce = astral.NewNonce()

	p, ok := s.pings.Set(nonce, &Ping{
		sentAt: time.Now(),
		pong:   make(chan struct{}),
	})
	if !ok {
		return -1, errors.New("duplicate sessionId")
	}
	defer s.pings.Delete(nonce)

	err := s.Write(&frames.Ping{
		Nonce: nonce,
	})
	if err != nil {
		return -1, err
	}
	p.sentAt = time.Now()

	select {
	case <-p.pong:
		return time.Since(p.sentAt), nil
	case <-time.After(s.pingTimeout):
		return -1, errors.New("ping timeout")
	case <-s.Stream.Done():
		return -1, s.Err()
	}
}

func (s *Link) pong(nonce astral.Nonce) (time.Duration, error) {
	p, ok := s.pings.Delete(nonce)
	if !ok {
		return -1, errors.New("invalid sessionId")
	}
	d := time.Since(p.sentAt)
	close(p.pong)
	if s.pressure != nil {
		s.pressure.OnRTT(d, time.Now())
	}
	return d, nil
}

// Wake triggers a ping on the next loop iteration.
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

func (s *Link) pingLoop() {
	select {
	case <-s.Stream.Done():
		return
	case <-time.After(time.Duration(rand.Int63n(int64(pingJitter)))):
	}

	for {
		if s.checks.Load() > 0 {
			select {
			case <-s.Stream.Done():
				return
			case <-time.After(activeInterval):
			}
		} else {
			select {
			case <-s.Stream.Done():
				return
			case <-s.wakeCh:
			}
		}

		if s.Err() != nil {
			return
		}

		if _, err := s.Ping(); err != nil {
			s.CloseWithError(err)
			return
		}

		if s.checks.Load() > 0 {
			s.checks.Add(-1)
		}
	}
}

func (s *Link) GetMux() *Mux {
	return s.mux
}
