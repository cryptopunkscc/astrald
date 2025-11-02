package nat

import (
	"bytes"
	"context"
	"encoding/binary"
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
)

// pairEntry represents a NAT pair runtime wrapper with keepalive and state.
type pairEntry struct {
	nat.EndpointPair
	state            atomic.Int32
	isPinger         bool
	lastPong         atomic.Int64
	keepaliveRunning atomic.Bool
	pings            sig.Map[astral.Nonce, int64] // nonce -> sentAt (unix nano)
	lockStarted      atomic.Int64                 // unix nano (for timeout)
	localIdentity    *astral.Identity             // which peer this entry represents
	conn             *net.UDPConn                 // UDP socket for raw packets
}

func (e *pairEntry) isIdle() bool    { return e.state.Load() == stateIdle }
func (e *pairEntry) isLocked() bool  { return e.state.Load() == stateLocked }
func (e *pairEntry) isExpired() bool { return e.state.Load() == stateExpired }

func (e *pairEntry) matchesPeer(peer *astral.Identity) bool {
	return e.PeerA.Identity.IsEqual(peer) || e.PeerB.Identity.IsEqual(peer)
}

func (e *pairEntry) init(localIdentity *astral.Identity, isPinger bool) error {
	e.localIdentity = localIdentity
	e.isPinger = isPinger
	e.state.Store(stateIdle)
	e.lastPong.Store(time.Now().UnixNano())

	localAddr := e.getLocalAddr()
	conn, err := net.ListenUDP("udp", localAddr)
	if err != nil {
		return err
	}
	e.conn = conn

	go e.recvLoop()

	if isPinger {
		e.startKeepalive()
	}
	return nil
}

func (e *pairEntry) lock() bool {
	if !e.state.CompareAndSwap(stateIdle, stateLocked) {
		return false
	}
	e.stopKeepalive()
	return true
}

func (e *pairEntry) expire() {
	switch e.state.Load() {
	case stateLocked:
		return
	default:
		e.state.Store(stateExpired)
		e.stopKeepalive()
		if c := e.conn; c != nil {
			e.conn = nil
			_ = c.Close()
		}

		for _, nonce := range e.pings.Keys() {
			e.pings.Delete(nonce)
		}
	}
}

func (e *pairEntry) startKeepalive() {
	if !e.keepaliveRunning.CompareAndSwap(false, true) {
		return
	}
	go e.keepaliveLoop()
}

func (e *pairEntry) stopKeepalive() {
	e.keepaliveRunning.Store(false)
}

func (e *pairEntry) keepaliveLoop() {
	defer e.keepaliveRunning.Store(false)
	ticker := time.NewTicker(pingInterval)
	defer ticker.Stop()

	for e.keepaliveRunning.Load() {
		switch e.state.Load() {
		case stateIdle:
			e.expirePings()
			// expire if no pong in a while (only for pinger)
			if e.isPinger && time.Since(time.Unix(0, e.lastPong.Load())) > pongTimeout {
				e.expire()
				return
			}
			if e.isPinger && e.sendPing() != nil {
				e.expire()
				return
			}

		case stateInLocking:
			e.expirePings()
			if e.isDrained() {
				_ = e.finalizeLock() // CAS-safe
				return
			}

			if e.lockTimedOut() {
				e.expire()
				return
			}

		default:
			return // Locked, InUse, Expired → silence
		}
		<-ticker.C
	}
}

// --- Ping / Pong ---
func (e *pairEntry) sendPing() error {
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
	remoteAddr := e.getRemoteAddr()
	_, err := c.WriteToUDP(buf.Bytes(), remoteAddr)
	return err
}

func (e *pairEntry) sendPong(nonce astral.Nonce) error {
	pong := &pingFrame{Nonce: nonce, Pong: true}

	var buf bytes.Buffer
	if _, err := pong.WriteTo(&buf); err != nil {
		return err
	}

	c := e.conn
	if c == nil {
		return io.ErrClosedPipe
	}
	remoteAddr := e.getRemoteAddr()
	_, err := c.WriteToUDP(buf.Bytes(), remoteAddr)
	return err
}

func (e *pairEntry) handlePing(nonce astral.Nonce, pong bool) {
	if e.state.Load() != stateIdle {
		return // silent in non-idle states
	}
	if pong {
		e.handlePong(nonce)
	} else {
		_ = e.sendPong(nonce)
	}
}

func (e *pairEntry) handlePong(nonce astral.Nonce) {
	if e.state.Load() != stateIdle {
		return
	}
	if _, ok := e.pings.Delete(nonce); ok {
		e.lastPong.Store(time.Now().UnixNano())
	}
}

func (e *pairEntry) expirePings() {
	now := time.Now()
	for _, nonce := range e.pings.Keys() {
		if sentAt, ok := e.pings.Get(nonce); ok {
			if now.Sub(time.Unix(0, sentAt)) > pingLifespan {
				e.pings.Delete(nonce)
			}
		}
	}
}

func (e *pairEntry) beginLock() bool {
	if !e.state.CompareAndSwap(stateIdle, stateInLocking) {
		return false
	}
	e.isPinger = false
	e.lockStarted.Store(time.Now().UnixNano())
	return true
}

func (e *pairEntry) finalizeLock() bool {
	if !e.state.CompareAndSwap(stateInLocking, stateLocked) {
		return false
	}
	for _, nonce := range e.pings.Keys() {
		e.pings.Delete(nonce)
	}
	// close UDP socket to enforce strict silence and unblock receiver
	if c := e.conn; c != nil {
		e.conn = nil
		_ = c.Close()
	}
	return true
}

func (e *pairEntry) isDrained() bool {
	return e.pings.Len() == 0
}

func (e *pairEntry) lockTimedOut() bool {
	start := e.lockStarted.Load()
	return start > 0 && time.Since(time.Unix(0, start)) > lockTimeout
}

func (e *pairEntry) getLocalAddr() *net.UDPAddr {
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

func (e *pairEntry) getRemoteAddr() *net.UDPAddr {
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

func (e *pairEntry) recvLoop() {
	buf := make([]byte, 64)
	for {
		c := e.conn
		if c == nil {
			return
		}

		n, _, err := c.ReadFromUDP(buf)
		// exit on socket error
		if err != nil {
			return
		}

		var f pingFrame
		if _, err := f.ReadFrom(bytes.NewReader(buf[:n])); err != nil {
			continue
		}
		e.handlePing(f.Nonce, f.Pong)
	}
}

// waitLocked blocks until the entry reaches stateLocked or the context is done.
// It also actively finalizes the lock when possible to avoid relying on keepalive.
func (e *pairEntry) waitLocked(ctx context.Context) error {
	// fast path
	if e.isLocked() {
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
		if e.state.Load() == stateInLocking && e.isDrained() {
			_ = e.finalizeLock()
		}
		if e.isLocked() {
			return nil
		}
		if e.lockTimedOut() {
			e.expire()
			return context.DeadlineExceeded
		}
		time.Sleep(backoff)
		if backoff < 100*time.Millisecond {
			backoff += 10 * time.Millisecond
		}
	}
}

type pingFrame struct {
	Nonce astral.Nonce
	Pong  bool
}

func (f *pingFrame) WriteTo(w io.Writer) (n int64, err error) {
	if err := binary.Write(w, binary.BigEndian, f.Nonce); err != nil {
		return 0, err
	}
	var pongByte byte
	if f.Pong {
		pongByte = 1
	}
	if err := binary.Write(w, binary.BigEndian, pongByte); err != nil {
		return 0, err
	}
	return 10, nil
}

func (f *pingFrame) ReadFrom(r io.Reader) (n int64, err error) {
	if err := binary.Read(r, binary.BigEndian, &f.Nonce); err != nil {
		return 0, err
	}
	var pongByte byte
	if err := binary.Read(r, binary.BigEndian, &pongByte); err != nil {
		return 0, err
	}
	f.Pong = pongByte == 1
	return 10, nil
}
