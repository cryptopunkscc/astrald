// conn.go
package udp

import (
	"net"
	"time"

	"github.com/cryptopunkscc/astrald/mod/exonet"
	"github.com/cryptopunkscc/astrald/mod/udp"
)

// DatagramWriter is how Conn sends bytes to its peer.
type DatagramWriter interface {
	WriteDatagram(b []byte) error
}

// DatagramReceiver is how Conn *receives* parsed packets when it does not own a socket read loop.
// (For active conns, the recvLoop calls HandleDatagram itself.)
type DatagramReceiver interface {
	HandleDatagram(raw []byte) // fast path: parse + process (ACK/data)
}

type Handshaker interface {
	Handshake() error
}

type Fragmenter interface {
}

// Conn represents a reliable UDP connection
// Handshake/FIN are out of scope for this MVP; stream semantics only.
type Conn struct {
	// socket / addressing
	udpConn        *net.UDPConn
	localEndpoint  *udp.Endpoint
	remoteEndpoint *udp.Endpoint
	// config
	cfg FlowControlConfig

	state      ConnState
	inCh       chan *Packet
	closedFlag bool

	//
	initialSeqNumLocal  uint32
	initialSeqNumRemote uint32
	// send state
	nextSeqNum  uint32
	connID      uint32
	sendBase    uint32 // oldest unacked sequence (i.e., cumulative ACK floor).
	ackedSeqNum uint32 // highest cumulative ACK seen (often == sendBase).
	expected    uint32
	//
	unacked map[uint32]*Packet // seq -> packet
	// receive state
}

func (c *Conn) setState(state ConnState) {
	c.state = state
}

func (c *Conn) inState(state ConnState) bool {
	return c.state == state
}

func (c *Conn) Read(p []byte) (n int, err error) {
	if !c.inState(StateEstablished) {
		return 0, udp.ErrConnectionNotEstablished
	}
	//TODO implement me
	panic("implement me")
}

func (c *Conn) Write(p []byte) (n int, err error) {
	if !c.inState(StateEstablished) {
		return 0, udp.ErrConnectionNotEstablished
	}

	//TODO implement me
	panic("implement me")
}

func (c *Conn) Close() error {
	c.closedFlag = true
	c.udpConn.SetReadDeadline(time.Now())
	//TODO implement me
	panic("implement me")
}

// NewConn constructs a connection around an already-connected UDP socket.
func NewConn(c *net.UDPConn, l, r *udp.Endpoint, cfg FlowControlConfig) (*Conn, error) {
	cfg.Normalize()
	if cfg.MSS <= 0 {
		return nil, udp.ErrZeroMSS
	}

	rc := &Conn{
		udpConn:        c,
		localEndpoint:  l,
		remoteEndpoint: r,
		cfg:            cfg,
	}

	return rc, nil
}

// Outbound reports whether this connection was dialed out.
// For now this always returns true for Dial usage; adjust if you add a listener.
func (c *Conn) Outbound() bool { return true }

// LocalEndpoint returns the local UDP endpoint.
func (c *Conn) LocalEndpoint() exonet.Endpoint {
	return c.localEndpoint
}

// RemoteEndpoint returns the remote UDP endpoint.
func (c *Conn) RemoteEndpoint() exonet.Endpoint {
	return c.remoteEndpoint
}

func (c *Conn) receivingLoop() {
	const maxPayloadSize = 64 * 1024
	buf := make([]byte, maxPayloadSize)
	for {
		c.udpConn.SetReadDeadline(time.Now().Add(1 * time.Second))
		n, addr, err := c.udpConn.ReadFromUDP(buf)
		if err != nil {
			if c.closedFlag {
				return
			}
			continue
		}

		// NOTE: test it
		if addr.String() != c.remoteEndpoint.IP.String() {
			continue // not for this Conn
		}

		pktData := make([]byte, n)
		copy(pktData, buf[:n])
		pkt := &Packet{}
		if err := pkt.Unmarshal(pktData); err != nil {
			continue // drop malformed
		}
		if int(pkt.Len) > maxPayloadSize {
			continue // invalid length
		}
		isControl := pkt.Flags&(FlagSYN|FlagACK|FlagFIN) != 0 && pkt.Len == 0
		if isControl {
			// Block until enqueued
			c.inCh <- pkt
		} else {
			// Drop data if channel full
			select {
			case c.inCh <- pkt:
			default:
				// drop data
			}
		}
	}
}

// Go
func (c *Conn) InboundPacketHandler() {
	for pkt := range c.inCh {
		if pkt.Flags&FlagACK != 0 {
			c.handleAckPacket(pkt)
			continue
		}

		if pkt.Flags&(FlagSYN|FlagFIN) != 0 {
			c.handleControlPacket(pkt)
			continue
		}

		c.handleDataPacket(pkt)
	}
}

// handleAckPacket processes ACK packets
func (c *Conn) handleAckPacket(pkt *Packet) {
	ack := pkt.Ack
	for seq := range c.unacked {
		if seq <= ack {
			delete(c.unacked, seq)
		}
	}
	c.sendBase = ack + 1
}

// handleControlPacket processes SYN, FIN, and other control packets
func (c *Conn) handleControlPacket(pkt *Packet) {
	// Example: handle SYN, FIN, or other control logic
	if pkt.Flags&FlagSYN != 0 {
		// ...handle SYN logic...
	}
	if pkt.Flags&FlagFIN != 0 {
		// ...handle FIN logic...
	}
	// ...handle other control flags as needed...
}

// handleDataPacket processes data packets
func (c *Conn) handleDataPacket(pkt *Packet) {
	// Example: deliver to receive buffer, update expected, send ACK, etc.
	// ...implement data delivery and ACK logic...
}
