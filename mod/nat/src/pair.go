package nat

import (
	"bytes"
	"context"
	"fmt"
	"net"
	"sync/atomic"
	"time"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/nat"
	"github.com/cryptopunkscc/astrald/sig"
)

const (
	defaultPingInterval  = 1 * time.Second
	defaultNoPingTimeout = 3 * time.Second  // if pinger doesn't get pong → Expire
	defaultPingLifespan  = 6 * time.Second  // drop stuck pings
	defaultLockTimeout   = 10 * time.Second // bound for locking handshake
	defaultMaxPingFails  = 10               // max writes before expiring
)

// Pair represents a NAT-traversed UDP port pair with keepalive and locking mechanisms.
// pair while being Idle makes sure to keep the NAT mapping alive by sending periodic pings.
// If the pair is inLocking state, it waits for pongs from the remote peer before completing the lock.
// Once locked, the pair closes its socket.
type Pair struct {
	nat.TraversedPortPair
	// configuration
	localIdentity *astral.Identity
	isPinger      bool
	onPairExpire  OnPairExpire // called when the pair expires

	conn net.PacketConn // usage of net.PacketConn interface allows us to write unit tests

	// State machine
	state     atomic.Int32 // nat.PairState
	lockStart atomic.Int64 // unix nano
	lastPing  atomic.Int64 // unix nano
	pings     sig.Map[astral.Nonce, int64]

	// Channels (safe goroutine communication)
	pingEvents chan pingEvent // receiver → runLoop
	wakeCh     chan struct{}  // wakes runLoop for state transitions
	lockedCh   chan struct{}  // closes when locked/expired

	// configuration options
	pingInterval  time.Duration
	noPingTimeout time.Duration
	pingLifespan  time.Duration
	lockTimeout   time.Duration
	maxPingFails  int
}

// Event Type from Receiver → runLoop
type pingEvent struct {
	nonce astral.Nonce
	pong  bool
}

// NewPair creates a new NAT pair with a bound UDP socket and applies the given options.
func NewPair(pair nat.TraversedPortPair, localID *astral.Identity, isPinger bool, opts ...PairOption) (*Pair, error) {
	conn, err := net.ListenUDP("udp", pair.GetLocalAddr(localID))
	if err != nil {
		return nil, err
	}

	return newPair(pair, localID, isPinger, conn, opts...), nil
}

// NewPairWithConn creates a pair with an injected PacketConn (for testing)
func NewPairWithConn(pair nat.TraversedPortPair, localID *astral.Identity, isPinger bool, conn net.PacketConn, opts ...PairOption) *Pair {
	return newPair(pair, localID, isPinger, conn, opts...)
}

// newPair initializes a new Pair with the given parameters and applies options.
func newPair(pair nat.TraversedPortPair, localID *astral.Identity, isPinger bool, conn net.PacketConn, opts ...PairOption) *Pair {
	p := &Pair{
		TraversedPortPair: pair,
		localIdentity:     localID,
		isPinger:          isPinger,
		conn:              conn,
		pingEvents:        make(chan pingEvent, 64),
		wakeCh:            make(chan struct{}, 1),
		lockedCh:          make(chan struct{}),
		pings:             sig.Map[astral.Nonce, int64]{},

		// default configuration options
		pingInterval:  defaultPingInterval,
		noPingTimeout: defaultNoPingTimeout,
		pingLifespan:  defaultPingLifespan,
		lockTimeout:   defaultLockTimeout,
		maxPingFails:  defaultMaxPingFails,
	}

	for _, opt := range opts {
		opt(p)
	}

	p.state.Store(int32(nat.StateIdle))
	p.lastPing.Store(time.Now().UnixNano())

	return p
}

// isExpectedAddr checks if the given UDP address matches the expected remote address
func (p *Pair) isExpectedAddr(addr *net.UDPAddr) bool {
	expected := p.GetRemoteAddr(p.localIdentity)
	return addr.IP.Equal(expected.IP) && addr.Port == expected.Port
}

// StartKeepAlive starts the receiver and run goroutines for keepalive management
func (p *Pair) StartKeepAlive(ctx context.Context) error {
	go p.receiver()
	go p.run(ctx)

	return nil
}

// receiver listens for UDP packets, validates addresses, parses frames, and queues events
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

// run executes the state machine loop, managing keepalive, locking, and expiration
func (p *Pair) run(ctx context.Context) {
	ticker := time.NewTicker(p.pingInterval)
	defer ticker.Stop()

	failCount := 0

	for {
		p.drainEvents()

		switch nat.PairState(p.state.Load()) {
		case nat.StateIdle:
			p.expirePings()

			if time.Since(time.Unix(0, p.lastPing.Load())) > p.noPingTimeout {
				p.Expire("no ping received for " + p.noPingTimeout.String() + "")
				return
			}

			if p.isPinger {
				if err := p.sendPing(); err != nil {
					failCount++
					if failCount >= p.maxPingFails {
						p.Expire("sending ping failed too many times")
						return
					}
				} else {
					failCount = 0
				}
			}
		case nat.StateInLocking:
			p.expirePings()

			if p.pings.Len() == 0 {
				p.finalizeLock()
				return
			}

			if p.lockTimedOut() {
				p.Expire("lock timed out")
				return
			}

		case nat.StateLocked, nat.StateExpired:
			return
		}

		select {
		case <-ctx.Done():
			p.Expire("context canceled")
			return
		case <-ticker.C:
		case <-p.wakeCh:
		}
	}
}

// drainEvents processes all queued ping events synchronously
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

// handlePing processes a ping event, updating timestamps and responding to pings
func (p *Pair) handlePing(ev pingEvent) {
	if nat.PairState(p.state.Load()) >= nat.StateLocked { // Locked or Expired
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

// sendPing sends a ping frame to the remote peer and tracks it
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

// sendPong sends a pong frame in response to a received ping
func (p *Pair) sendPong(nonce astral.Nonce) error {
	f := pingFrame{Nonce: nonce, Pong: true}
	var buf bytes.Buffer
	if _, err := f.WriteTo(&buf); err != nil {
		return err
	}
	_, err := p.conn.WriteTo(buf.Bytes(), p.GetRemoteAddr(p.localIdentity))
	return err
}

// expirePings removes pings that have exceeded their configured lifespan
func (p *Pair) expirePings() {
	now := time.Now().UnixNano()
	clone := p.pings.Clone()
	for nonce, ts := range clone {
		if now-ts > p.pingLifespan.Nanoseconds() {
			p.pings.Delete(nonce)
		}
	}
}

// lockTimedOut checks if the locking process has exceeded the timeout
func (p *Pair) lockTimedOut() bool {
	return time.Since(time.Unix(0, p.lockStart.Load())) > p.lockTimeout
}

// BeginLock initiates locking if the pair is idle, returning success status
func (p *Pair) BeginLock() bool {
	if !p.state.CompareAndSwap(int32(nat.StateIdle), int32(nat.StateInLocking)) {
		return false
	}

	p.lockStart.Store(time.Now().UnixNano())
	p.wake()
	return true
}

// finalizeLock completes locking by closing the connection and notifying waiters
func (p *Pair) finalizeLock() bool {
	if !p.state.CompareAndSwap(int32(nat.StateInLocking), int32(nat.StateLocked)) {
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

// Expire transitions to expired state, cleans up resources, and notifies
func (p *Pair) Expire(reason string) {
	p.state.Store(int32(nat.StateExpired))
	_ = p.conn.Close()

	clone := p.pings.Clone()
	for n := range clone {
		p.pings.Delete(n)
	}

	select {
	case <-p.lockedCh:
	default:
		close(p.lockedCh)
	}

	if p.onPairExpire != nil {
		fmt.Println("REASON FOR PAIR EXPIRATION: ", reason)
		p.onPairExpire(p)
	}

	p.wake()
}

// WaitLocked waits for the pair to lock or the context to cancel
func (p *Pair) WaitLocked(ctx context.Context) error {
	if nat.PairState(p.state.Load()) == nat.StateLocked {
		return nil
	}

	p.wake()

	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-p.lockedCh:
		if nat.PairState(p.state.Load()) == nat.StateLocked {
			return nil
		}
		return nat.ErrPairCantLock
	}
}

// State returns the current state of the pair
func (p *Pair) State() nat.PairState {
	return nat.PairState(p.state.Load())
}

// IsIdle returns true if the pair is in idle state
func (p *Pair) IsIdle() bool {
	return nat.PairState(p.state.Load()) == nat.StateIdle
}

// IsExpired returns true if the pair is expired
func (p *Pair) IsExpired() bool {
	return nat.PairState(p.state.Load()) == nat.StateExpired
}

// IsLocked returns true if the pair is locked
func (p *Pair) IsLocked() bool {
	return nat.PairState(p.state.Load()) == nat.StateLocked
}

// LastPing returns the timestamp of the last received ping/pong
func (p *Pair) LastPing() time.Time {
	return time.Unix(0, p.lastPing.Load())
}

// LockTimeout returns the maximum duration for locking handshake
func (p *Pair) LockTimeout() time.Duration {
	return p.lockTimeout
}

// wake sends a non-blocking wake signal to the run loop
func (p *Pair) wake() {
	select {
	case p.wakeCh <- struct{}{}:
	default:
	}
}

// Options
type OnPairExpire func(p *Pair)

type PairOption func(*Pair)

func WithPingInterval(d time.Duration) PairOption {
	return func(p *Pair) { p.pingInterval = d }
}

func WithNoPingTimeout(d time.Duration) PairOption {
	return func(p *Pair) { p.noPingTimeout = d }
}

func WithPingLifespan(d time.Duration) PairOption {
	return func(p *Pair) { p.pingLifespan = d }
}

func WithLockTimeout(d time.Duration) PairOption {
	return func(p *Pair) { p.lockTimeout = d }
}

func WithMaxPingFails(n int) PairOption {
	return func(p *Pair) { p.maxPingFails = n }
}

func WithOnPairExpire(f OnPairExpire) PairOption {
	return func(p *Pair) { p.onPairExpire = f }
}
