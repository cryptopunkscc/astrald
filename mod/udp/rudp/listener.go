package rudp

import (
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/udp"
)

const defaultAcceptBacklog = 32
const defaultHandshakeTimeout = 5 * time.Second

// Listener accepts inbound RUDP connections (server side) and returns only
// fully established connections via Accept().
type Listener struct {
	udpConn  *net.UDPConn
	cfg      Config
	baseCtx  *astral.Context
	mu       sync.Mutex
	conns    map[string]*Conn // remoteKey -> Conn (includes handshaking ones)
	acceptCh chan *Conn
	closed   atomic.Bool
}

// Listen creates a new Listener bound to addr. handshakeTimeout==0 falls back to a small default.
func Listen(ctx *astral.Context, addr *net.UDPAddr, cfg Config, handshakeTimeout time.Duration) (*Listener, error) {
	conn, err := net.ListenUDP("udp", addr)
	if err != nil {
		return nil, err
	}
	cfg.Normalize()
	if handshakeTimeout <= 0 {
		handshakeTimeout = defaultHandshakeTimeout
	}
	l := &Listener{
		udpConn:  conn,
		cfg:      cfg,
		baseCtx:  ctx,
		conns:    make(map[string]*Conn),
		acceptCh: make(chan *Conn, defaultAcceptBacklog),
	}
	go l.readLoop(handshakeTimeout)
	return l, nil
}

// Addr returns the underlying listening address.
func (l *Listener) Addr() net.Addr { return l.udpConn.LocalAddr() }

// Accept blocks until an established connection is available, the context is canceled, or the listener is closed.
func (l *Listener) Accept(ctx *astral.Context) (*Conn, error) {
	for {
		if l.closed.Load() {
			return nil, udp.ErrListenerClosed
		}
		select {
		case c, ok := <-l.acceptCh:
			if !ok {
				return nil, udp.ErrListenerClosed
			}
			return c, nil
		case <-ctx.Done():
			return nil, ctx.Err()
		}
	}
}

// Close shuts down the listener and all active connections.
func (l *Listener) Close() error {
	if !l.closed.CompareAndSwap(false, true) {
		return udp.ErrListenerClosed
	}
	// Closing the UDP socket unblocks readLoop.
	_ = l.udpConn.Close()
	l.mu.Lock()
	for _, c := range l.conns {
		c.Close()
	}
	l.conns = nil
	l.mu.Unlock()
	close(l.acceptCh)
	return nil
}

// readLoop performs demultiplexing and inbound connection setup.
func (l *Listener) readLoop(handshakeTimeout time.Duration) {
	buf := make([]byte, 64*1024)
	for {
		if l.closed.Load() {
			return
		}
		n, addr, err := l.udpConn.ReadFromUDP(buf)
		if err != nil {
			if l.closed.Load() {
				return
			}
			continue
		}
		if n < 13 { // minimal header length
			continue
		}
		pkt := &Packet{}
		if err := pkt.Unmarshal(buf[:n]); err != nil {
			continue
		}

		remoteKey := addr.String()
		l.mu.Lock()
		conn := l.conns[remoteKey]
		if conn == nil && pkt.Flags&FlagSYN != 0 { // new inbound attempt
			remoteEP, perr := udp.ParseEndpoint(addr.String())
			if perr != nil {
				l.mu.Unlock()
				continue
			}
			localEP, _ := udp.ParseEndpoint(l.udpConn.LocalAddr().String())
			// Per-handshake timeout context derived from baseCtx
			hCtx, cancel := l.baseCtx.WithTimeout(handshakeTimeout)
			c, cerr := NewConn(l.udpConn, localEP, remoteEP, l.cfg, false, pkt, hCtx)
			if cerr != nil { // immediate constructor error; cancel context
				cancel()
				l.mu.Unlock()
				continue
			}
			c.OnEstablished(func(ec *Conn) {
				if l.closed.Load() {
					cancel()
					return
				}
				select {
				case l.acceptCh <- ec:
				default:
				}
				cancel()
			})
			c.OnClosed(func(ec *Conn) {
				l.mu.Lock()
				delete(l.conns, remoteKey)
				l.mu.Unlock()
				cancel()
			})
			l.conns[remoteKey] = c
			conn = c
		}
		l.mu.Unlock()

		if conn != nil {
			conn.ProcessPacket(pkt)
		}
	}
}
