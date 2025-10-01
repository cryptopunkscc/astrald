// conn.go
package rudp

import (
	"errors"
	"fmt"
	"io"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/exonet"
	"github.com/cryptopunkscc/astrald/mod/udp"
	"github.com/smallnest/ringbuffer"
)

type Unacked struct {
	pkt         *Packet   // Packet metadata (seq, len)
	sentTime    time.Time // Last sent time
	rtxCount    int       // Retransmit count
	length      int       // Payload length
	isHandshake bool      // True if this entry is for a handshake control packet
}

// Conn represents a reliable UDP connection.
// Implements reliability, flow control, retransmissions, and error notification.
// Key mechanisms:
//   - Reliable delivery using retransmission timer and fast retransmit
//   - Flow control using packet window (MaxWindowPackets)
//   - Concurrency safety via sendMu and sendCond
//   - Error notification via ErrChan (application can monitor for connection-level errors)
//   - Centralized resource cleanup via Close()
//   - PoC limitations: no congestion control, no SACK, no adaptive pacing
type Conn struct {
	// UDP socket and addressing
	udpConn         *net.UDPConn // Underlying UDP socket
	localEndpoint   *udp.Endpoint
	remoteEndpoint  *udp.Endpoint
	outbound        bool // true if we initiated the connection
	onEstablishedCb func(*Conn)
	onClosedCb      func(*Conn)

	// Configuration (reliability, flow control, etc.)
	cfg Config // All protocol parameters

	// Connection state (atomic for lock-free reads)
	state      uint32       // Current connection state (stores ConnState)
	inCh       chan *Packet // Incoming packet channel
	closedFlag uint32       // 0=open, 1=closed (atomic)

	// Sequence numbers and send state
	initialSeqNumLocal  uint32
	initialSeqNumRemote uint32
	nextSeqNum          uint32
	connID              uint32 // Connection ID
	sendBase            uint32 // Oldest unacked sequence (ACK floor)
	ackedSeqNum         uint32 // Highest cumulative ACK seen
	expected            uint32 // Next expected sequence number (receive side)

	inflight uint32 // Number of unacked packets

	// Send buffer and reliability
	sendRB  *ringbuffer.RingBuffer // Persistent send ring buffer (FIFO; bytes consumed at packetization)
	frag    *BasicFragmenter       // Fragmenter for packetization
	unacked map[uint32]*Unacked    // Map of unacked packets (seq -> Unacked); stores full packet copies for retransmission

	// Concurrency and coordination
	sendMu   sync.Mutex // Protects all shared state
	sendCond *sync.Cond // Condition variable for sender coordination

	// Retransmission timer
	rtxTimer *time.Timer // Fixed retransmission timer (PoC only)

	// Error notification
	ErrChan chan error // Channel for connection-level errors (e.g., retransmission failure)

	// Inbound buffering & ACK state
	recvRB      *ringbuffer.RingBuffer
	recvMu      sync.Mutex
	recvCond    *sync.Cond
	ackTimer    *time.Timer
	ackPending  bool
	lastAckSent uint32
	// Out-of-order buffer (keyed by sequence number of first byte)
	recvOO map[uint32]*Packet // stored packets with Seq > expected awaiting in-order delivery
}

// unified handshake channel capacity (applies to inbound & outbound)
const handshakeQueueCap = 64

// NewConn constructs a connection around a UDP socket.
// Parameters:
//
//	outbound   - true for client initiated connection (owns socket & recv loop)
//	firstPacket / ctx - required only for inbound (outbound=false) to drive server handshake
func NewConn(cn *net.UDPConn, l, r *udp.Endpoint, cfg Config, outbound bool, firstPacket *Packet, ctx *astral.Context) (*Conn, error) {
	cfg.Normalize()
	if cfg.MaxSegmentSize <= 0 {
		return nil, udp.ErrZeroMSS
	}

	if !outbound {
		if firstPacket == nil || ctx == nil {
			return nil, errors.New("inbound connection requires firstPacket and ctx")
		}
	}

	sendRBSize := cfg.MaxWindowBytes * 2 // allow for some retransmit slack
	rb := ringbuffer.New(sendRBSize)
	frag := NewBasicFragmenter(cfg.MaxSegmentSize)

	rc := &Conn{
		udpConn:        cn,
		localEndpoint:  l,
		remoteEndpoint: r,
		cfg:            cfg,
		sendRB:         rb,
		frag:           frag,
		unacked:        make(map[uint32]*Unacked),
		ErrChan:        make(chan error, 1),                   // Buffered to avoid blocking
		inCh:           make(chan *Packet, handshakeQueueCap), // handshake delivery channel
		outbound:       outbound,
	}
	rc.sendCond = sync.NewCond(&rc.sendMu)
	rc.recvRB = ringbuffer.New(cfg.RecvBufBytes)
	rc.recvCond = sync.NewCond(&rc.recvMu)
	rc.recvOO = make(map[uint32]*Packet)

	if outbound {
		// Do not start recvLoop here; StartClientHandshake will start it after handshake completes
	} else {
		go rc.recvLoop()
	}

	if !outbound {
		// Start server handshake asynchronously for inbound connections
		go func() {
			if err := rc.StartServerHandshake(ctx, firstPacket); err != nil {
				rc.Close()
			}
		}()
	}

	return rc, nil
}

// OnEstablished registers a callback invoked exactly once when the connection transitions to Established.
func (c *Conn) OnEstablished(cb func(*Conn)) {
	c.sendMu.Lock()
	c.onEstablishedCb = cb
	c.sendMu.Unlock()
	// Fast path: if already established, invoke asynchronously
	if c.inState(StateEstablished) && cb != nil {
		go cb(c)
	}
}

// OnClosed registers a callback invoked exactly once after Close() releases resources.
// If the connection is already closed when registering, the callback is invoked asynchronously.
func (c *Conn) OnClosed(cb func(*Conn)) {
	c.sendMu.Lock()
	c.onClosedCb = cb
	closed := c.isClosed()
	c.sendMu.Unlock()
	if closed && cb != nil {
		go cb(c)
	}
}

func (c *Conn) setState(state ConnState) {
	atomic.StoreUint32(&c.state, uint32(state))
}

func (c *Conn) inState(state ConnState) bool {
	return atomic.LoadUint32(&c.state) == uint32(state)
}

func (c *Conn) isClosed() bool { return atomic.LoadUint32(&c.closedFlag) != 0 }

func (c *Conn) Read(p []byte) (n int, err error) {
	if !c.inState(StateEstablished) && !c.isClosed() {
		return 0, udp.ErrConnectionNotEstablished
	}
	c.recvMu.Lock()
	for c.recvRB != nil && c.recvRB.Length() == 0 && !c.isClosed() {
		c.recvCond.Wait()
	}
	if c.recvRB == nil || (c.recvRB.Length() == 0 && c.isClosed()) {
		c.recvMu.Unlock()
		return 0, io.EOF
	}
	want := len(p)
	if rl := int(c.recvRB.Length()); want > rl {
		want = rl
	}
	if want == 0 {
		c.recvMu.Unlock()
		return 0, nil
	}
	// Read directly into caller's buffer (no temp allocation)
	m, _ := c.recvRB.Read(p[:want])
	c.recvMu.Unlock()
	return m, nil
}

// Write enqueues data into the send ring buffer. Implementation in send.go.
func (c *Conn) Write(p []byte) (n int, err error) {
	return c.writeSend(p)
}

func (c *Conn) Close() error {
	c.sendMu.Lock()
	if c.isClosed() { // already closed
		c.sendMu.Unlock()
		return nil
	}
	atomic.StoreUint32(&c.closedFlag, 1)
	// stop retransmission timer if running
	if c.rtxTimer != nil {
		c.rtxTimer.Stop()
		c.rtxTimer = nil
	}
	// wake any waiters (writers, senderLoop)
	c.sendCond.Broadcast()
	ch := c.inCh
	c.inCh = nil // detach channel to prevent further sends
	closedCb := c.onClosedCb
	c.sendMu.Unlock()

	if ch != nil {
		close(ch)
	}
	if c.outbound { // only needed to unblock recvLoop for outbound
		_ = c.udpConn.SetReadDeadline(time.Now())
	}
	c.recvMu.Lock()
	if c.ackTimer != nil {
		c.ackTimer.Stop()
		c.ackTimer = nil
	}
	if c.recvCond != nil {
		c.recvCond.Broadcast()
	}
	c.recvMu.Unlock()
	var err error
	if c.outbound {
		err = c.udpConn.Close()
	}
	// Invoke close callback after resources released
	if closedCb != nil {
		closedCb(c)
	}

	return err
}

func (c *Conn) onEstablished() {
	// Idempotent: only transition once
	if c.inState(StateEstablished) || c.isClosed() {
		return
	}
	// Initialize receive-side expected sequence to remote initial seq + 1 (account for SYN consuming one seq)
	if c.initialSeqNumRemote != 0 && c.expected == 0 {
		c.expected = c.initialSeqNumRemote + 1
	}
	c.setState(StateEstablished)
	// Future established-only initializations (keepalives, metrics, etc.) go here.
	go c.senderLoop()
	// Invoke callback (outside locks) if set
	cb := func() func(*Conn) { c.sendMu.Lock(); defer c.sendMu.Unlock(); return c.onEstablishedCb }()
	if cb != nil {
		cb(c)
	}
}

// HandleAckPacket processes ACK packets
func (c *Conn) HandleAckPacket(packet *Packet) {
	ack := packet.Ack
	c.sendMu.Lock()
	defer c.sendMu.Unlock()
	if ack > c.ackedSeqNum {
		c.ackedSeqNum = ack
	}
	if ack > c.sendBase {
		c.sendBase = ack
	}
	// Remove fully acked packets (keyed by seq)
	for s, u := range c.unacked {
		if u.isHandshake {
			// Handshake control (SYN / SYN|ACK) conceptually consumes 1 sequence number.
			// Require ack > s (i.e., ack == s+1) to delete, avoiding premature removal
			// if an unexpected ack echo with ack==s arrives.
			if ack > s { // expected ack == s+1
				delete(c.unacked, s)
			}
			continue
		}
		// Data packet: remove when cumulative ack covers entire payload.
		if s+uint32(u.length) <= ack {
			delete(c.unacked, s)
		}
	}
	// Stop retransmission timer if no unacked packets remain
	if len(c.unacked) == 0 && c.rtxTimer != nil {
		c.rtxTimer.Stop()
		c.rtxTimer = nil
	}
	c.sendCond.Broadcast()
}

// HandleControlPacket processes SYN, FIN, and other control packets
func (c *Conn) HandleControlPacket(packet *Packet) {
	// Example: handle SYN, FIN, or other control logic
	if packet.Flags&FlagSYN != 0 {
		// TODO: ...handle SYN logic...
	}
	if packet.Flags&FlagFIN != 0 {
		// TODO: ...handle FIN logic...
	}
	// ...handle other control flags as needed...
}

// Interface compliance for exonet.Conn
func (c *Conn) Outbound() bool { return c.outbound }
func (c *Conn) LocalEndpoint() exonet.Endpoint {
	if c == nil {
		return nil
	}
	return c.localEndpoint
}
func (c *Conn) RemoteEndpoint() exonet.Endpoint {
	if c == nil {
		return nil
	}
	return c.remoteEndpoint
}

// ProcessPacket feeds a received packet into the connection (server-side demux path).
// If the connection is not yet established, the packet is queued for handshake processing.
func (c *Conn) ProcessPacket(pkt *Packet) {
	if !c.inState(StateEstablished) {
		c.sendMu.Lock()
		ch := c.inCh
		c.sendMu.Unlock()
		if ch != nil {
			select {
			case ch <- pkt:
			default:
			}
		}
		return
	}
	if pkt.Flags&FlagACK != 0 {
		c.HandleAckPacket(pkt)
		return
	}
	if pkt.Flags&(FlagSYN|FlagFIN) != 0 {
		c.HandleControlPacket(pkt)
		return
	}
	c.handleDataPacket(pkt)
}

// sendDatagram sends a raw packet buffer choosing the correct syscall based on
// connection role (outbound connections use Write on a connected socket;
// inbound connections use WriteToUDP specifying the remote endpoint).
func (c *Conn) sendDatagram(b []byte) (n int, err error) {
	if c.udpConn == nil {
		return n, udp.ErrConnClosed
	}
	if c.outbound {
		_, err := c.udpConn.Write(b)
		return n, err
	}
	if c.remoteEndpoint == nil {
		return n, fmt.Errorf("remote endpoint nil")
	}
	written, err := c.udpConn.WriteToUDP(b, c.remoteEndpoint.UDPAddr())
	return written, err
}
