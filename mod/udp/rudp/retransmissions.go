package rudp

import (
	"sort"
	"time"

	"github.com/cryptopunkscc/astrald/mod/udp"
)

// queueAckLocked marks that an ACK should be (re)sent. Caller must hold recvMu.
func (c *Conn) queueAckLocked() { c.ackPending = true }

// triggerAck decides immediate vs delayed ACK after recvMu has been released.
func (c *Conn) triggerAck(ackDelay time.Duration) {
	if ackDelay == 0 {
		c.sendPureACK()
	} else {
		c.scheduleAck()
	}
}

// scheduleAck sets / resets delayed ACK timer
func (c *Conn) scheduleAck() {
	c.recvMu.Lock()
	if !c.ackPending || c.isClosed() {
		c.recvMu.Unlock()
		return
	}
	d := c.cfg.AckDelay
	if d <= 0 {
		ackNeeded := c.ackPending
		c.ackPending = false
		c.recvMu.Unlock()
		if ackNeeded {
			c.sendPureACK()
		}
		return
	}
	if c.ackTimer != nil {
		c.ackTimer.Reset(d)
	} else {
		c.ackTimer = time.AfterFunc(d, c.fireAck)
	}
	c.recvMu.Unlock()
}

func (c *Conn) fireAck() {
	c.recvMu.Lock()
	if !c.ackPending || c.isClosed() {
		c.recvMu.Unlock()
		return
	}
	c.ackPending = false
	c.recvMu.Unlock()
	c.sendPureACK()
}

// sendPureACK sends a standalone ACK reflecting current expected sequence.
func (c *Conn) sendPureACK() {
	if !c.inState(StateEstablished) || c.isClosed() {
		return
	}
	// snapshot expected & window
	c.recvMu.Lock()
	exp := c.expected
	winFree := uint32(0)
	if c.recvRB != nil {
		winFree = uint32(c.recvRB.Free())
	}
	c.recvMu.Unlock()
	// clamp window to 16-bit
	win := uint16(0)
	if winFree > 0xFFFF {
		win = 0xFFFF
	} else {
		win = uint16(winFree)
	}
	pkt := &Packet{Seq: 0, Ack: exp, Flags: FlagACK, Win: win, Len: 0}
	b, err := pkt.Marshal()
	if err != nil {
		return
	}
	// best-effort send via unified path
	_, _ = c.sendDatagram(b)
}

// handleRetransmissionTimeoutLocked assumes sendMu is held and performs retransmissions.
// Returns true if the retransmission limit was exceeded for any packet.
func (c *Conn) handleRetransmissionTimeoutLocked() (limitExceeded bool) {
	if len(c.unacked) == 0 {
		return false
	}

	seqs := make([]uint32, 0, len(c.unacked))
	for s := range c.unacked {
		seqs = append(seqs, s)
	}
	sort.Slice(seqs, func(i, j int) bool { return seqs[i] < seqs[j] })

	for _, s := range seqs {
		u := c.unacked[s]
		if u.rtxCount >= c.cfg.RetransmissionLimit {
			limitExceeded = true
			break
		}

		// Update ACK field to latest cumulative ACK and retransmit
		u.pkt.Ack = c.ackedSeqNum
		b, err := u.pkt.Marshal()
		if err == nil {
			_, _ = c.sendDatagram(b)
		}
		u.rtxCount++
		u.sentTime = time.Now()
	}
	return
}

// handleRetransmissionTimeout provides backward-compatible external behavior (locking & close semantics)
func (c *Conn) handleRetransmissionTimeout() {
	c.sendMu.Lock()
	limitExceeded := c.handleRetransmissionTimeoutLocked()
	c.sendMu.Unlock()
	if limitExceeded {
		select {
		case c.ErrChan <- udp.ErrRetransmissionLimitExceeded:
		default:
		}
		c.Close()
		close(c.ErrChan)
	}
}
