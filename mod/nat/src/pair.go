package nat

import (
	"bytes"
	"context"
	"io"
	"net"
	"sync/atomic"
	"time"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/nat"
	"github.com/cryptopunkscc/astrald/sig"
)

const (
	stateIdle = iota
	stateInLocking
	stateLocked
	stateExpired
)

const (
	pingInterval = 1 * time.Second
	pongTimeout  = 3 * time.Second
	pingLifespan = 6 * time.Second  // remove old ping if not ponged
	lockTimeout  = 10 * time.Second // safety bound for InLocking
	maxPingFails = 10
)

// Pair represents a NAT Pair runtime wrapper with keepalive and state.
type Pair struct {
	nat.TraversedPortPair
	state            atomic.Int32
	isPinger         bool
	lastPong         atomic.Int64
	keepaliveRunning atomic.Bool
	pings            sig.Map[astral.Nonce, int64] // nonce -> sentAt (unix nano)
	lockStarted      atomic.Int64                 // unix nano (for timeout)
	localIdentity    *astral.Identity             // which peer this entry represents
	conn             *net.UDPConn                 // UDP socket for raw packets
}

func (e *Pair) IsIdle() bool    { return e.state.Load() == stateIdle }
func (e *Pair) IsLocked() bool  { return e.state.Load() == stateLocked }
func (e *Pair) IsExpired() bool { return e.state.Load() == stateExpired }
func (e *Pair) IsDrained() bool {
	return e.pings.Len() == 0
}

func (e *Pair) Lock() bool {
	if !e.state.CompareAndSwap(stateIdle, stateLocked) {
		return false
	}

	e.stopKeepAlive()
	return true
}

func (e *Pair) Expire() {
	switch e.state.Load() {
	case stateLocked:
		return
	default:
		e.state.Store(stateExpired)
		e.stopKeepAlive()
		if c := e.conn; c != nil {
			_ = c.Close()
		}

		for _, nonce := range e.pings.Keys() {
			e.pings.Delete(nonce)
		}
	}
}

func (e *Pair) MatchesPeer(peer *astral.Identity) bool {
	return e.PeerA.Identity.IsEqual(peer) || e.PeerB.Identity.IsEqual(peer)
}

func (e *Pair) StartKeepAlive(localIdentity *astral.Identity, isPinger bool) error {
	e.localIdentity = localIdentity
	e.isPinger = isPinger
	e.state.Store(stateIdle)
	e.lastPong.Store(time.Now().UnixNano())

	localAddr := e.GetLocalAddr()
	conn, err := net.ListenUDP("udp", localAddr)
	if err != nil {
		return err
	}
	e.conn = conn

	go e.pingReceiver()
	if isPinger {
		e.startKeepAlive()
	}
	return nil
}

func (e *Pair) startKeepAlive() {
	if !e.keepaliveRunning.CompareAndSwap(false, true) {
		return
	}
	go e.keepAliveLoop()
}

func (e *Pair) stopKeepAlive() {
	e.keepaliveRunning.Store(false)
}

func (e *Pair) keepAliveLoop() {
	defer e.keepaliveRunning.Store(false)
	ticker := time.NewTicker(pingInterval)
	defer ticker.Stop()
	failCount := 0

	for e.keepaliveRunning.Load() {
		switch e.state.Load() {
		case stateIdle:
			e.expirePings()
			// Expire if no pong in a while (only for pinger)
			if e.isPinger && time.Since(time.Unix(0, e.lastPong.Load())) > pongTimeout {
				e.Expire()
				return
			}

			if e.isPinger {
				if err := e.sendPing(); err != nil {
					failCount++
					if failCount >= maxPingFails {
						e.Expire()
						return
					}
				} else {
					failCount = 0
				}
			}

		case stateInLocking:
			e.expirePings()
			if e.IsDrained() {
				_ = e.FinalizeLock()
				return
			}

			if e.lockTimedOut() {
				e.Expire()
				return
			}

		default:
			return // Locked, InUse, Expired → silence
		}
		<-ticker.C
	}
}

func (e *Pair) sendPing() error {
	nonce := astral.NewNonce()
	e.pings.Set(nonce, time.Now().UnixNano())

	ping := &pingFrame{Nonce: nonce}
	var buf bytes.Buffer
	if _, err := ping.WriteTo(&buf); err != nil {
		return err
	}

	c := e.conn
	if c == nil {
		return io.ErrClosedPipe
	}
	remoteAddr := e.GetRemoteAddr()
	_, err := c.WriteToUDP(buf.Bytes(), remoteAddr)
	return err
}

func (e *Pair) sendPong(nonce astral.Nonce) error {
	pong := &pingFrame{Nonce: nonce, Pong: true}

	var buf bytes.Buffer
	if _, err := pong.WriteTo(&buf); err != nil {
		return err
	}

	c := e.conn
	if c == nil {
		return io.ErrClosedPipe
	}
	remoteAddr := e.GetRemoteAddr()
	_, err := c.WriteToUDP(buf.Bytes(), remoteAddr)
	return err
}

func (e *Pair) handlePing(nonce astral.Nonce, pong bool) {
	if e.state.Load() != stateIdle {
		return // silent in non-idle states
	}
	if pong {
		e.handlePong(nonce)
	} else {
		_ = e.sendPong(nonce)
	}
}

func (e *Pair) handlePong(nonce astral.Nonce) {
	if e.state.Load() != stateIdle {
		return
	}
	if _, ok := e.pings.Delete(nonce); ok {
		e.lastPong.Store(time.Now().UnixNano())
	}
}

func (e *Pair) expirePings() {
	now := time.Now()
	for _, nonce := range e.pings.Keys() {
		if sentAt, ok := e.pings.Get(nonce); ok {
			if now.Sub(time.Unix(0, sentAt)) > pingLifespan {
				e.pings.Delete(nonce)
			}
		}
	}
}

func (e *Pair) BeginLock() bool {
	if !e.state.CompareAndSwap(stateIdle, stateInLocking) {
		return false
	}
	e.lockStarted.Store(time.Now().UnixNano())
	return true
}

func (e *Pair) FinalizeLock() bool {
	if !e.state.CompareAndSwap(stateInLocking, stateLocked) {
		return false
	}
	for _, nonce := range e.pings.Keys() {
		e.pings.Delete(nonce)
	}
	// close UDP socket to enforce strict silence and unblock receiver
	if c := e.conn; c != nil {
		_ = c.Close()
	}
	return true
}

func (e *Pair) lockTimedOut() bool {
	start := e.lockStarted.Load()
	return start > 0 && time.Since(time.Unix(0, start)) > lockTimeout
}

func (e *Pair) GetLocalAddr() *net.UDPAddr {
	if e.localIdentity.IsEqual(e.PeerA.Identity) {
		return &net.UDPAddr{
			IP:   net.IP(e.PeerA.Endpoint.IP),
			Port: int(e.PeerA.Endpoint.Port),
		}
	}
	return &net.UDPAddr{
		IP:   net.IP(e.PeerB.Endpoint.IP),
		Port: int(e.PeerB.Endpoint.Port),
	}
}

func (e *Pair) GetRemoteAddr() *net.UDPAddr {
	if e.localIdentity.IsEqual(e.PeerA.Identity) {
		return &net.UDPAddr{
			IP:   net.IP(e.PeerB.Endpoint.IP),
			Port: int(e.PeerB.Endpoint.Port),
		}
	}
	return &net.UDPAddr{
		IP:   net.IP(e.PeerA.Endpoint.IP),
		Port: int(e.PeerA.Endpoint.Port),
	}
}

func (e *Pair) pingReceiver() {
	buf := make([]byte, 64)
	for {
		c := e.conn
		if c == nil {
			return
		}

		n, _, err := c.ReadFromUDP(buf)
		if err != nil {
			// FIXME:
			return
		}

		var f pingFrame
		if _, err := f.ReadFrom(bytes.NewReader(buf[:n])); err != nil {
			continue
		}
		e.handlePing(f.Nonce, f.Pong)
	}
}

// WaitLocked blocks until the entry reaches stateLocked or the context is done.
// It also actively finalizes the Lock when possible to avoid relying on keepalive.
func (e *Pair) WaitLocked(ctx context.Context) error {
	// fast path
	if e.IsLocked() {
		return nil
	}

	backoff := 10 * time.Millisecond
	for {
		select {
		case <-ctx.Done():
			return ctx.Err()
		default:
		}

		// If we're in locking and the ping buffer is drained, finalize now.
		if e.state.Load() == stateInLocking && e.IsDrained() {
			ok := e.FinalizeLock()
			if !ok {
				return nat.ErrPairCantLock
			}

		}
		if e.IsLocked() {
			return nil
		}
		if e.lockTimedOut() {
			e.Expire()
			return context.DeadlineExceeded
		}
		time.Sleep(backoff)
		if backoff < 100*time.Millisecond {
			backoff += 10 * time.Millisecond
		}
	}
}
