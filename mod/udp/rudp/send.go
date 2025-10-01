package rudp

import (
	"time"
)

// Write enqueues data into the send ring buffer. It blocks until enough space is available.
// Signals the sender goroutine after enqueue.
func (c *Conn) writeSend(p []byte) (n int, err error) {
	c.sendMu.Lock()
	defer c.sendMu.Unlock()
	writeLen := len(p)
	for writeLen > int(c.sendRB.Free()) {
		c.sendCond.Wait()
	}
	_, err = c.sendRB.Write(p)
	if err != nil {
		return 0, err
	}
	// Only the senderLoop needs to be awakened; writers waiting for space are not helped by a write.
	c.sendCond.Signal()
	return writeLen, nil
}

// windowFull returns true if the packet window is full (no more packets can be sent).
func (c *Conn) windowFull() bool {
	return len(c.unacked) >= c.cfg.MaxWindowPackets
}

// commitPacketLocked registers the packet as unacked and advances sequence numbers. Caller holds sendMu.
func (c *Conn) commitPacketLocked(pkt *Packet) (startTimer bool) {
	seq := pkt.Seq
	c.unacked[seq] = &Unacked{pkt: pkt, sentTime: time.Now(), rtxCount: 0, length: int(pkt.Len)}
	c.nextSeqNum += uint32(pkt.Len)
	return len(c.unacked) == 1
}

// armRetransmitTimer starts retransmission timer if needed (no lock held).
func (c *Conn) armRetransmitTimer(need bool) {
	if need {
		c.startRtxTimer()
	}
}

// senderLoop runs as a goroutine and is responsible for sending packets from the send buffer.
// Inlined fragmentation & packet build: we avoid an extra lock/unlock by performing
// (select bytes -> copy -> build packet -> commit) under one critical section, then
// releasing the lock before marshaling and sending.
// FIFO model: bytes are consumed from sendRB as soon as they are packetized. Retransmissions
// use copies stored in unacked map. No random access over the ring is performed.
func (c *Conn) senderLoop() {
	defer func() { c.sendCond.Broadcast() }()
	for {
		c.sendMu.Lock()
		for {
			if c.isClosed() || !c.inState(StateEstablished) {
				c.sendMu.Unlock()
				return
			}
			if c.sendRB.Length() > 0 && !c.windowFull() {
				break
			}
			c.sendCond.Wait()
		}

		// Decide fragment size and drain payload while still holding the lock.
		ask := int(c.sendRB.Length())
		if ask > c.cfg.MaxSegmentSize {
			ask = c.cfg.MaxSegmentSize
		}
		// Allocate buffer and read from ring (destructive read).
		payload := make([]byte, ask)
		readN, _ := c.sendRB.Read(payload)
		if readN == 0 { // nothing actually read; loop and re-evaluate predicates
			c.sendMu.Unlock()
			continue
		}
		if readN != ask { // shrink payload if ring gave us fewer bytes
			payload = payload[:readN]
		}

		seq := c.nextSeqNum
		pkt := &Packet{Seq: seq, Ack: c.ackedSeqNum, Flags: FlagACK, Len: uint16(len(payload)), Payload: payload}
		startTimer := c.commitPacketLocked(pkt)
		// We freed ring space; signal at least one waiting writer or (rarely) another waiter.
		c.sendCond.Signal()
		c.sendMu.Unlock()

		// Manage retransmission timer outside lock
		c.armRetransmitTimer(startTimer)

		// Marshal & send outside lock to minimize critical section time.
		raw, err := pkt.Marshal()
		if err != nil {
			select {
			case c.ErrChan <- err:
			default:
			}
			c.Close()
			return
		}
		if _, err := c.sendDatagram(raw); err != nil {
			// Suppress error (packet retained in unacked for retransmission). Continue loop.
			continue
		}
		// next iteration
	}
}
