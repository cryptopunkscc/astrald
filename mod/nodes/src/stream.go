package nodes

import (
	"errors"
	"fmt"
	"math/rand"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/mod/exonet"
	"github.com/cryptopunkscc/astrald/mod/nodes/frames"
	"github.com/cryptopunkscc/astrald/sig"
)

type Link struct {
	ch          *channel.Channel
	id          astral.Nonce
	createdAt   time.Time
	conn        astral.Conn
	pings       sig.Map[astral.Nonce, *Ping]
	checks      atomic.Int32
	throughput  atomic.Uint64
	outbound    bool
	pingTimeout time.Duration
	wakeCh      chan struct{}
	pressure    StreamPressureDetector

	mu   sync.Mutex
	err  sig.Value[error]
	done chan struct{}
	in   chan frames.Frame // fixme: remove
}

type Ping struct {
	sentAt time.Time
	pong   chan struct{}
}

func newLink(conn astral.Conn, id astral.Nonce, outbound bool) *Link {
	s := &Link{
		id:          id,
		conn:        conn,
		createdAt:   time.Now(),
		outbound:    outbound,
		pingTimeout: defaultPingTimeout,
		wakeCh:      make(chan struct{}, 1),
		done:        make(chan struct{}),
		in:          make(chan frames.Frame, 32),
		ch:          channel.New(conn),
	}

	go s.reader()
	go s.pingLoop()

	return s
}

func (s *Link) reader() {
	var rerr error
	defer func() {
		s.err.Swap(nil, rerr)
		_ = s.ch.Close()
		close(s.in)
		close(s.done)
	}()

	for {
		obj, err := s.ch.Receive()
		if err != nil {
			rerr = err
			return
		}
		frame, ok := obj.(frames.Frame)
		if !ok {
			rerr = fmt.Errorf("decoded object is not a Frame: %T", obj)
			return
		}
		s.in <- frame
	}
}

func (s *Link) Read() <-chan frames.Frame { return s.in }

func (s *Link) Done() <-chan struct{} { return s.done }

func (s *Link) Err() error { return s.err.Get() }

func (s *Link) LocalIdentity() *astral.Identity {
	return s.conn.LocalIdentity()
}

func (s *Link) RemoteIdentity() *astral.Identity {
	return s.conn.RemoteIdentity()
}

func (s *Link) CloseWithError(err error) error {
	if err == nil {
		err = errors.New("link closed")
	}
	s.err.Swap(nil, err)
	_ = s.ch.Close()
	return nil
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

func (s *Link) PressureHigh() bool {
	return s.pressure != nil && s.pressure.IsHigh()
}

func (s *Link) Throughput() uint64 {
	return s.throughput.Load()
}

func (s *Link) Write(frame frames.Frame) error {
	if f, ok := frame.(*frames.Data); ok {
		s.throughput.Add(uint64(len(f.Payload)))
		if s.pressure != nil {
			s.pressure.OnBytes(len(f.Payload), time.Now())
		}
	}

	if _, ok := frame.(*frames.Ping); !ok {
		s.check()
	}

	s.mu.Lock()
	defer s.mu.Unlock()

	if err := s.ch.Send(frame); err != nil {
		s.err.Swap(nil, err)
		_ = s.ch.Close()
		return err
	}
	return nil
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
	case <-s.done:
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

		if _, err := s.Ping(); err != nil {
			s.CloseWithError(err)
			return
		}

		if s.checks.Load() > 0 {
			s.checks.Add(-1)
		}
	}
}
