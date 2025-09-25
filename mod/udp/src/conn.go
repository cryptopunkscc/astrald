// conn.go
package udp

import (
	"bytes"
	"context"
	"io"
	"net"
	"sync"
	"sync/atomic"
	"time"

	"github.com/cryptopunkscc/astrald/mod/exonet"
	"github.com/cryptopunkscc/astrald/mod/udp"
)

// Conn implements src, ordered communication over a connected UDP socket.
// Handshake/FIN are out of scope for this MVP; stream semantics only.
type Conn struct {
	// socket / addressing
	udpConn        *net.UDPConn
	localEndpoint  *udp.Endpoint
	remoteEndpoint *udp.Endpoint

	// config
	cfg FlowControlConfig
	mss int

	// send side (guarded by sendMu unless noted)
	sendMu        sync.Mutex
	sendBase      uint32 // first unacked byte
	nextSeq       uint32 // next byte sequence to assign
	nextSendSeq   uint32
	sendQ         *bytes.Buffer // queued app data (bounded by cfg.SendBufBytes)
	sendCond      *sync.Cond    // signals space available / data added
	bytesInFlight int
	unacked       map[uint32]segMeta // seqStart -> meta
	order         []uint32           // seqStarts in send order (oldest first)

	// recv side
	rcvMu      sync.Mutex
	rcvNext    uint32
	ooo        map[uint32][]byte // out-of-order segments by seqStart
	appBuf     *ringBuffer       // ordered bytes for Read()
	ackPending atomic.Bool       // (reserved) if you add explicit flags later
	// timers
	rtoMu    sync.Mutex
	rto      time.Duration
	rtoTimer *time.Timer
	ackTimer *time.Timer // set on-demand in recv.go

	// control/lifecycle
	ctx    context.Context
	cancel context.CancelFunc
	wg     sync.WaitGroup

	closed    atomic.Bool
	closeOnce sync.Once
	closeErr  atomic.Value // error

	// write serialization (shared per UDP socket if you ever share it)
	writeMu *sync.Mutex
	// Add a mutex field for synchronization
	mutex sync.Mutex
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
		mss:            cfg.MSS,

		sendBase: 1, // start at 1 so 0 can be a sentinel in traces
		nextSeq:  1,
		sendQ:    &bytes.Buffer{},
		unacked:  make(map[uint32]segMeta),
		order:    make([]uint32, 0, 128),

		rcvNext: 1,
		ooo:     make(map[uint32][]byte),
		appBuf:  newRingBuffer(cfg.RecvBufBytes),

		rto:     cfg.RTO,
		writeMu: &sync.Mutex{},
	}
	rc.sendCond = sync.NewCond(&rc.sendMu)

	// Start receiver loop (defined in recv.go)
	rc.wg.Add(1)
	go rc.recvLoop()

	return rc, nil
}

// Read implements stream semantics. It blocks until data is available or the
// connection is closed and drained. On close, it returns any stored terminal error
// or io.EOF when the buffer is empty.
func (c *Conn) Read(p []byte) (int, error) {
	n, err := c.appBuf.Read(p)
	if n > 0 {
		return n, nil
	}
	if c.closed.Load() {
		if errv := c.closeErr.Load(); errv != nil {
			return 0, errv.(error)
		}
		return 0, io.EOF
	}
	return n, err
}

// Close terminates the connection and waits for the recv loop to exit.
func (c *Conn) Close() error {
	c.closeOnce.Do(func() {
		c.closed.Store(true)
		c.cancel()

		// stop timers
		c.stopRTO() // defined in send.go
		c.rtoMu.Lock()
		if c.ackTimer != nil {
			c.ackTimer.Stop()
			c.ackTimer = nil
		}
		c.rtoMu.Unlock()

		// wake blocked goroutines
		c.sendCond.Broadcast()
		c.appBuf.Close()

		_ = c.udpConn.Close()
	})
	c.wg.Wait()
	return nil
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

// NOTE: Flagged for check (might be ai overcomplexity)
// closeWithError records the error and closes the connection.
func (c *Conn) closeWithError(err error) {
	if err != nil {
		c.closeErr.Store(err)
	}
	_ = c.Close()
}

// seqLT compares sequence numbers with wrap-around semantics.
func seqLT(a, b uint32) bool { return int32(a-b) < 0 }

// sendPacket sends a packet over the UDP connection.
func (c *Conn) sendPacket(pkt *Packet) error {
	raw, err := pkt.Marshal()
	if err != nil {
		return err
	}

	c.writeMu.Lock()
	defer c.writeMu.Unlock()

	_, err = c.udpConn.Write(raw)
	return err
}

// handleRTO handles retransmission timeouts by retransmitting the earliest unacked segment.
func (c *Conn) handleRTO() {
	// Implementation for retransmission timeout handling
	// This will involve retransmitting the earliest unacked segment
	// and applying exponential backoff to the retransmission timer.
}
