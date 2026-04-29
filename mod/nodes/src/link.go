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
	"github.com/cryptopunkscc/astrald/mod/nodes"
	"github.com/cryptopunkscc/astrald/mod/nodes/frames"
	"github.com/cryptopunkscc/astrald/sig"
)

var _ nodes.Link = &Link{}
var _ nodes.QualityLink = &Link{}
var _ nodes.NetworkLink = &Link{}

type Link struct {
	*channel.Channel
	id             astral.Nonce
	createdAt      time.Time
	localIdentity  *astral.Identity
	remoteIdentity *astral.Identity
	pings          sig.Map[astral.Nonce, *Ping]
	checks         atomic.Int32
	throughput     atomic.Uint64
	outbound       bool
	localEp        exonet.Endpoint
	remoteEp       exonet.Endpoint
	pingTimeout    time.Duration
	wakeCh         chan struct{}
	pressure       LinkPressureDetector

	mu   sync.Mutex
	err  sig.Value[error]
	done chan struct{}
	in   chan frames.Frame
}

type Ping struct {
	sentAt time.Time
	pong   chan struct{}
}

func newLink(ch *channel.Channel, localIdentity, remoteIdentity *astral.Identity, id astral.Nonce, outbound bool, localEp, remoteEp exonet.Endpoint) *Link {
	s := &Link{
		Channel:        ch,
		id:             id,
		localIdentity:  localIdentity,
		remoteIdentity: remoteIdentity,
		createdAt:      time.Now(),
		outbound:       outbound,
		localEp:        localEp,
		remoteEp:       remoteEp,
		pingTimeout:    defaultPingTimeout,
		wakeCh:         make(chan struct{}, 1),
		done:           make(chan struct{}),
		in:             make(chan frames.Frame, 32),
	}

	go s.reader()
	go s.pingLoop()

	return s
}

func (s *Link) reader() {
	var rerr error
	defer func() {
		s.err.Swap(nil, rerr)
		_ = s.Channel.Close()
		close(s.in)
		close(s.done)
	}()

	for {
		obj, err := s.Receive()
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
	return s.localIdentity
}

func (s *Link) RemoteIdentity() *astral.Identity {
	return s.remoteIdentity
}

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

func (s *Link) LocalEndpoint() exonet.Endpoint {
	return s.localEp
}

func (s *Link) RemoteEndpoint() exonet.Endpoint {
	return s.remoteEp
}

// Wake triggers a ping on the next loop iteration.
func (s *Link) Wake() {
	select {
	case s.wakeCh <- struct{}{}:
	default:
	}
}

func (s *Link) IsHighPressure() bool {
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

	if err := s.Send(frame); err != nil {
		s.err.Swap(nil, err)
		_ = s.Channel.Close()
		return err
	}
	return nil
}

func (s *Link) ping() (time.Duration, error) {
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

		if _, err := s.ping(); err != nil {
			s.CloseWithError(err)
			return
		}

		if s.checks.Load() > 0 {
			s.checks.Add(-1)
		}
	}
}
