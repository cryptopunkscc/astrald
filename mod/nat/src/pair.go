package nat

import (
	"bytes"
	"context"
	"net"
	"sync/atomic"
	"time"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/nat"
	"github.com/cryptopunkscc/astrald/sig"
)

const (
	pingInterval  = 1 * time.Second
	noPingTimeout = 3 * time.Second  // if pinger doesn't get pong → expire
	pingLifespan  = 6 * time.Second  // drop stuck pings
	lockTimeout   = 10 * time.Second // bound for locking handshake
	maxPingFails  = 10               // max writes before expiring
)

// State Machine Values
type PairState int32

const (
	StateIdle      PairState = iota // normal keepalive
	StateInLocking                  // lock requested, waiting for drain
	StateLocked                     // socket silent, no traffic
	StateExpired                    // permanently closed
)

type Pair struct {
	nat.TraversedPortPair
	// configuration
	localIdentity *astral.Identity
	isPinger      bool

	// socket (owned by runLoop)
	conn net.PacketConn

	// State machine
	state     atomic.Int32 // PairState
	lockStart atomic.Int64 // unix nano
	lastPing  atomic.Int64 // unix nano (pinger only)
	pings     sig.Map[astral.Nonce, int64]

	// Channels (safe goroutine communication)
	pingEvents chan pingEvent // receiver → runLoop
	wakeCh     chan struct{}  // wakes runLoop for state transitions
	lockedCh   chan struct{}  // closes when locked/expired
}

// Event Type from Receiver → runLoop
type pingEvent struct {
	nonce astral.Nonce
	pong  bool
}

// NewPair creates a pair that auto-binds a UDP socket
func NewPair(pair nat.TraversedPortPair, localID *astral.Identity, isPinger bool) (*Pair, error) {
	conn, err := net.ListenUDP("udp", pair.GetLocalAddr(localID))
	if err != nil {
		return nil, err
	}

	return newPair(pair, localID, isPinger, conn), nil
}

// NewPairWithConn creates a pair with an injected PacketConn (for testing)
func NewPairWithConn(pair nat.TraversedPortPair, localID *astral.Identity, isPinger bool, conn net.PacketConn) *Pair {
	return newPair(pair, localID, isPinger, conn)
}

// newPair is the internal constructor that sets up all fields
func newPair(pair nat.TraversedPortPair, localID *astral.Identity, isPinger bool, conn net.PacketConn) *Pair {
	p := &Pair{
		TraversedPortPair: pair,
		localIdentity:     localID,
		isPinger:          isPinger,
		conn:              conn,
		pingEvents:        make(chan pingEvent, 64),
		wakeCh:            make(chan struct{}, 1),
		lockedCh:          make(chan struct{}),
		pings:             sig.Map[astral.Nonce, int64]{},
	}

	p.state.Store(int32(StateIdle))
	p.lastPing.Store(time.Now().UnixNano())

	return p
}

// isExpectedAddr checks if the given UDP address matches the expected remote address
func (p *Pair) isExpectedAddr(addr *net.UDPAddr) bool {
	expected := p.GetRemoteAddr(p.localIdentity)
	return addr.IP.Equal(expected.IP) && addr.Port == expected.Port
}

func (p *Pair) StartKeepAlive(ctx context.Context) error {
	go p.receiver()
	go p.run(ctx)

	return nil
}

func (p *Pair) receiver() {
	buf := make([]byte, 512)

	for {
		n, addr, err := p.conn.ReadFrom(buf)
		if err != nil {
			return
		}

		remoteAddr, ok := addr.(*net.UDPAddr)
		if !ok || !p.isExpectedAddr(remoteAddr) {
			continue
		}

		var f pingFrame
		if _, err := f.ReadFrom(bytes.NewReader(buf[:n])); err != nil {
			continue
		}

		p.pingEvents <- pingEvent{nonce: f.Nonce, pong: f.Pong}
		p.wake()
	}
}
func (p *Pair) run(ctx context.Context) {
	ticker := time.NewTicker(pingInterval)
	defer ticker.Stop()

	failCount := 0

	for {
		p.drainEvents()

		switch PairState(p.state.Load()) {
		case StateIdle:
			p.expirePings()

			if time.Since(time.Unix(0, p.lastPing.Load())) > noPingTimeout {
				p.expire()
				return
			}

			if p.isPinger {
				if err := p.sendPing(); err != nil {
					failCount++
					if failCount >= maxPingFails {
						p.expire()
						return
					}
				} else {
					failCount = 0
				}
			}
		case StateInLocking:
			p.expirePings()

			if p.pings.Len() == 0 {
				p.finalizeLock()
				return
			}

			if p.lockTimedOut() {
				p.expire()
				return
			}

		case StateLocked, StateExpired:
			return
		}

		select {
		case <-ctx.Done():
			p.expire()
			return
		case <-ticker.C:
		case <-p.wakeCh:
		}
	}
}

func (p *Pair) drainEvents() {
	for {
		select {
		case ev := <-p.pingEvents:
			p.handlePing(ev)
		default:
			return
		}
	}
}

func (p *Pair) handlePing(ev pingEvent) {
	if PairState(p.state.Load()) >= StateLocked { // Locked or Expired
		return
	}

	p.lastPing.Store(time.Now().UnixNano())

	if ev.pong {
		if _, ok := p.pings.Delete(ev.nonce); ok {
			p.wake()
		}
	} else {
		_ = p.sendPong(ev.nonce)
	}
}

func (p *Pair) sendPing() error {
	nonce := astral.NewNonce()
	p.pings.Set(nonce, time.Now().UnixNano())
	f := pingFrame{Nonce: nonce, Pong: false}
	var buf bytes.Buffer
	if _, err := f.WriteTo(&buf); err != nil {
		return err
	}
	_, err := p.conn.WriteTo(buf.Bytes(), p.GetRemoteAddr(p.localIdentity))
	return err
}

func (p *Pair) sendPong(nonce astral.Nonce) error {
	f := pingFrame{Nonce: nonce, Pong: true}
	var buf bytes.Buffer
	if _, err := f.WriteTo(&buf); err != nil {
		return err
	}
	_, err := p.conn.WriteTo(buf.Bytes(), p.GetRemoteAddr(p.localIdentity))
	return err
}

func (p *Pair) expirePings() {
	now := time.Now().UnixNano()
	clone := p.pings.Clone()
	for nonce, ts := range clone {
		if now-ts > pingLifespan.Nanoseconds() {
			p.pings.Delete(nonce)
		}
	}
}

func (p *Pair) lockTimedOut() bool {
	return time.Since(time.Unix(0, p.lockStart.Load())) > lockTimeout
}

func (p *Pair) BeginLock() bool {
	if !p.state.CompareAndSwap(int32(StateIdle), int32(StateInLocking)) {
		return false
	}

	p.lockStart.Store(time.Now().UnixNano())
	p.wake()
	return true
}

func (p *Pair) finalizeLock() bool {
	if !p.state.CompareAndSwap(int32(StateInLocking), int32(StateLocked)) {
		return false
	}

	clone := p.pings.Clone()
	for n := range clone {
		p.pings.Delete(n)
	}

	if p.conn != nil {
		_ = p.conn.Close()
	}

	select {
	case <-p.lockedCh:
	default:
		close(p.lockedCh)
	}

	return true
}

func (p *Pair) Expire() {
	if p.state.Load() == int32(StateLocked) {
		return
	}
	p.state.Store(int32(StateExpired))
	p.wake()
}

func (p *Pair) expire() {
	p.state.Store(int32(StateExpired))
	if p.conn != nil {
		_ = p.conn.Close()
	}
	clone := p.pings.Clone()
	for n := range clone {
		p.pings.Delete(n)
	}

	select {
	case <-p.lockedCh:
	default:
		close(p.lockedCh)
	}
}

// WaitLocked()
func (p *Pair) WaitLocked(ctx context.Context) error {
	if PairState(p.state.Load()) == StateLocked {
		return nil
	}

	p.wake()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-p.lockedCh:
		if PairState(p.state.Load()) == StateLocked {
			return nil
		}
		return nat.ErrPairCantLock
	}
}

func (p *Pair) State() PairState {
	return PairState(p.state.Load())
}

func (p *Pair) IsIdle() bool {
	return PairState(p.state.Load()) == StateIdle
}

func (p *Pair) IsExpired() bool {
	return PairState(p.state.Load()) == StateExpired
}

func (p *Pair) IsLocked() bool {
	return PairState(p.state.Load()) == StateLocked
}

func (p *Pair) LastPing() time.Time {
	return time.Unix(0, p.lastPing.Load())
}

func (p *Pair) wake() {
	select {
	case p.wakeCh <- struct{}{}:
	default:
	}
}
