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
	c.sendCond.Broadcast()
	return writeLen, nil
}

// windowFull returns true if the packet window is full (no more packets can be sent).
func (c *Conn) windowFull() bool {
	return len(c.unacked) >= c.cfg.MaxWindowPackets
}

// planFragmentLocked decides how many bytes to send (<= ask) and drains them into a fresh buffer.
// Caller MUST hold sendMu. Returns nil buffer if nothing to send.
func (c *Conn) planFragmentLocked(ask int) (seq uint32, buf []byte, n int) {
	if ask <= 0 || c.sendRB.Length() == 0 {
		return 0, nil, 0
	}
	if ask > int(c.sendRB.Length()) {
		ask = int(c.sendRB.Length())
	}
	if ask > c.cfg.MaxSegmentSize {
		ask = c.cfg.MaxSegmentSize
	}
	fragBuf := make([]byte, ask)
	readN, _ := c.sendRB.Read(fragBuf)
	if readN == 0 {
		return 0, nil, 0
	}
	return c.nextSeqNum, fragBuf[:readN], readN
}

// buildPacket converts raw payload into a Packet and marshals it.
func (c *Conn) buildPacket(seq uint32, payload []byte) (*Packet, []byte, error) {
	pkt := &Packet{Seq: seq, Ack: c.ackedSeqNum, Flags: FlagACK, Len: uint16(len(payload)), Payload: payload}
	b, err := pkt.Marshal()
	if err != nil {
		return nil, nil, err
	}
	return pkt, b, nil
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

// rollbackPacketLocked removes an unacked packet on send failure. Caller does NOT hold lock upon entry.
func (c *Conn) rollbackPacketLocked(seq uint32, length int) {
	c.sendMu.Lock()
	if u, ok := c.unacked[seq]; ok && u.length == length {
		delete(c.unacked, seq)
		if c.nextSeqNum == seq+uint32(length) {
			c.nextSeqNum = seq
		}
		if len(c.unacked) == 0 && c.rtxTimer != nil {
			c.rtxTimer.Stop()
			c.rtxTimer = nil
		}
		c.sendCond.Broadcast()
	}
	c.sendMu.Unlock()
}

// sendFragment consumes up to ask bytes, builds a packet and sends it.
func (c *Conn) sendFragment(ask int) (bool, error) {
	c.sendMu.Lock()
	seq, payload, plen := c.planFragmentLocked(ask)
	if plen == 0 {
		c.sendMu.Unlock()
		return false, nil
	}
	pkt, raw, err := c.buildPacket(seq, payload)
	if err != nil {
		c.sendMu.Unlock()
		return false, err
	}
	startTimer := c.commitPacketLocked(pkt)
	c.sendCond.Broadcast()
	c.sendMu.Unlock()

	c.armRetransmitTimer(startTimer)
	if _, err := c.sendDatagram(raw); err != nil {
		c.rollbackPacketLocked(seq, int(pkt.Len))
		return false, err
	}
	return true, nil
}

// startRtxTimer arms the retransmission timer if not already running
func (c *Conn) startRtxTimer() {
	c.sendMu.Lock()
	if c.rtxTimer != nil {
		c.sendMu.Unlock()
		return
	}
	interval := c.cfg.RetransmissionInterval
	c.rtxTimer = time.AfterFunc(interval, func() {
		c.handleRetransmissionTimeout()
		c.sendMu.Lock()
		if len(c.unacked) > 0 && !c.isClosed() {
			c.rtxTimer.Reset(interval)
		} else {
			c.rtxTimer = nil
		}
		c.sendMu.Unlock()
	})
	c.sendMu.Unlock()
}

// senderLoop runs as a goroutine and is responsible for sending packets from the send buffer.
// FIFO model: bytes are consumed from sendRB as soon as they are packetized. Retransmissions
// use copies stored in unacked map. No random access over the ring is performed.
func (c *Conn) senderLoop() {
	defer func() { c.sendCond.Broadcast() }()
	for {
		c.sendMu.Lock()
		if c.isClosed() || !c.inState(StateEstablished) {
			c.sendMu.Unlock()
			return
		}
		for c.sendRB.Length() == 0 || c.windowFull() {
			c.sendCond.Wait()
			if c.isClosed() || !c.inState(StateEstablished) {
				c.sendMu.Unlock()
				return
			}
		}
		ask := int(c.sendRB.Length())
		if ask > c.cfg.MaxSegmentSize {
			ask = c.cfg.MaxSegmentSize
		}
		c.sendMu.Unlock()

		if sent, _ := c.sendFragment(ask); !sent {
			continue
		}
	}
}
