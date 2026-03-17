// nat/pair_test.go
package nat_test

import (
	"net"
	"sync"
	"time"
)

// ─────────────────────────────────────────────────────────────────────────────
// Perfect in-memory net.PacketConn pair — lives only in nat_test

type memPacketConn struct {
	mu     sync.Mutex
	peer   *memPacketConn
	queue  []packet
	closed bool
	local  net.Addr
	remote net.Addr
}

type packet struct {
	data []byte
	from net.Addr
}

// NewMemPacketPair creates two perfectly connected in-memory UDP-like sockets.
// No framing. No length prefix. No bugs. Exact UDP semantics.
func NewMemPacketPair() (net.PacketConn, net.PacketConn) {
	aLocal := &net.UDPAddr{IP: net.ParseIP("10.0.0.1"), Port: 40000}
	aRemote := &net.UDPAddr{IP: net.ParseIP("10.0.0.2"), Port: 40001}
	bLocal := &net.UDPAddr{IP: net.ParseIP("10.0.0.2"), Port: 40001}
	bRemote := &net.UDPAddr{IP: net.ParseIP("10.0.0.1"), Port: 40000}

	a := &memPacketConn{local: aLocal, remote: aRemote}
	b := &memPacketConn{local: bLocal, remote: bRemote}

	a.peer = b
	b.peer = a

	return a, b
}

// NewPipePacketPair — fully configurable, just like before
func NewPipePacketPair(
	localA, remoteA,
	localB, remoteB *net.UDPAddr,
) (net.PacketConn, net.PacketConn) {
	if localA == nil {
		localA = &net.UDPAddr{IP: net.ParseIP("10.0.0.1"), Port: 40000}
	}
	if remoteA == nil {
		remoteA = &net.UDPAddr{IP: net.ParseIP("10.0.0.2"), Port: 40001}
	}
	if localB == nil {
		localB = &net.UDPAddr{IP: net.ParseIP("10.0.0.2"), Port: 40001}
	}
	if remoteB == nil {
		remoteB = &net.UDPAddr{IP: net.ParseIP("10.0.0.1"), Port: 40000}
	}
	return NewMemPacketPairWithAddrs(localA, remoteA, localB, remoteB)
}

// Optional: fully configurable addresses (use when testing isExpectedAddr)
func NewMemPacketPairWithAddrs(
	aLocal, aRemote,
	bLocal, bRemote *net.UDPAddr,
) (net.PacketConn, net.PacketConn) {
	a := &memPacketConn{local: aLocal, remote: aRemote}
	b := &memPacketConn{local: bLocal, remote: bRemote}
	a.peer = b
	b.peer = a
	return a, b
}

// WriteTo sends a full datagram to the peer
func (c *memPacketConn) WriteTo(p []byte, _ net.Addr) (int, error) {
	c.mu.Lock()
	if c.closed {
		c.mu.Unlock()
		return 0, net.ErrClosed
	}
	data := append([]byte(nil), p...) // copy
	c.mu.Unlock()

	c.peer.mu.Lock()
	defer c.peer.mu.Unlock()
	if c.peer.closed {
		return 0, net.ErrClosed
	}
	c.peer.queue = append(c.peer.queue, packet{data: data, from: c.local})
	return len(p), nil
}

// ReadFrom blocks until a full packet arrives
func (c *memPacketConn) ReadFrom(p []byte) (int, net.Addr, error) {
	for {
		c.mu.Lock()
		if len(c.queue) > 0 {
			pkt := c.queue[0]
			c.queue = c.queue[1:]
			if len(c.queue) == 0 {
				c.queue = nil
			}
			c.mu.Unlock()
			n := copy(p, pkt.data)
			return n, pkt.from, nil
		}
		if c.closed {
			c.mu.Unlock()
			return 0, nil, net.ErrClosed
		}
		c.mu.Unlock()
		time.Sleep(time.Millisecond) // prevent CPU spin
	}
}

func (c *memPacketConn) Close() error {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.closed = true
	return nil
}

func (c *memPacketConn) LocalAddr() net.Addr                { return c.local }
func (c *memPacketConn) RemoteAddr() net.Addr               { return c.remote }
func (c *memPacketConn) SetDeadline(t time.Time) error      { return nil }
func (c *memPacketConn) SetReadDeadline(t time.Time) error  { return nil }
func (c *memPacketConn) SetWriteDeadline(t time.Time) error { return nil }

var _ net.PacketConn = &memPacketConn{}
