package udp

import (
	"fmt"
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

// nextSendOffset returns the offset in sendRB for the next fragment to send.
// If no packets are unacked, starts at sendBase. Otherwise, finds the highest offset of unacked packets.
func (c *Conn) nextSendOffset() int {
	if len(c.unacked) == 0 {
		return int(c.sendBase)
	}

	maxOff := 0
	for _, u := range c.unacked {
		if u.offset+u.length > maxOff {
			maxOff = u.offset + u.length
		}
	}
	return maxOff
}

// fillFragBuf fills buf with ask bytes from sendRB at offset off.
// Returns an error if the buffer is too small or the read fails.
func (c *Conn) fillFragBuf(buf []byte, _ int, ask int) error {
	if ask > len(buf) {
		return fmt.Errorf("fragBuf too small: ask=%d, buf=%d", ask, len(buf))
	}
	readN, err := c.sendRB.Read(buf[:ask])
	if err != nil {
		return err
	}
	if readN != ask {
		return fmt.Errorf("partial read from sendRB: expected %d, got %d", ask, readN)
	}
	return nil
}

// sendFragment fragments, marshals, and sends a packet from sendRB at offset off, up to ask bytes.
// Updates unacked and nextSeqNum. Returns true if a packet was sent, false otherwise.
func (c *Conn) sendFragment(off int, ask int) (bool, error) {
	c.sendMu.Lock()

	fragBuf := make([]byte, ask)
	if err := c.fillFragBuf(fragBuf, off, ask); err != nil {
		c.sendMu.Unlock()
		return false, err
	}

	pkt, pktLen, ok := c.frag.MakeNew(c.nextSeqNum, ask, &ByteStreamBuffer{data: fragBuf[:ask]})
	if !ok || pktLen == 0 {
		c.sendMu.Unlock()
		return false, nil
	}
	pkt.Ack = c.ackedSeqNum

	b, err := pkt.Marshal()
	if err != nil {
		c.sendMu.Unlock()
		return false, err
	}

	seq := c.nextSeqNum
	c.unacked[seq] = &Unacked{
		pkt:      pkt,
		sentTime: time.Now(),
		rtxCount: 0,
		offset:   off,
		length:   pktLen,
	}
	c.nextSeqNum += uint32(pktLen)
	startTimer := len(c.unacked) == 1
	c.sendCond.Broadcast() // wake writers: space freed by consumption
	c.sendMu.Unlock()

	if startTimer {
		c.startRtxTimer()
	}

	// Network I/O outside lock
	_, err = c.udpConn.WriteToUDP(b, c.remoteEndpoint.UDPAddr())
	if err != nil {
		c.sendMu.Lock()
		if u, ok2 := c.unacked[seq]; ok2 && u.length == pktLen {
			delete(c.unacked, seq)
			if c.nextSeqNum == seq+uint32(pktLen) { // rewind if no later sends
				c.nextSeqNum = seq
			}
			if len(c.unacked) == 0 && c.rtxTimer != nil {
				c.rtxTimer.Stop()
				c.rtxTimer = nil
			}
			c.sendCond.Broadcast() // notify writers rollback restored space
		}
		c.sendMu.Unlock()
		return false, err
	}

	return true, nil
}

// startRtxTimer arms the retransmission timer if not already running
func (c *Conn) startRtxTimer() {
	c.sendMu.Lock()
	if c.rtxTimer != nil {
		c.sendMu.Unlock()
		return // already running
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
// It enforces flow control (packet window), fragments data, and sends packets.
// The loop blocks when the window is full or the buffer is empty, and wakes up on sendCond.
// Exits cleanly on connection close or state change, and notifies waiters.
// For PoC, does not implement advanced pacing, batching, or prioritization.
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
		off := c.nextSendOffset()
		end := int(c.sendRB.Length())
		if off >= end { // nothing to send after all
			c.sendMu.Unlock()
			continue
		}
		rem := end - off
		ask := rem
		if ask > c.cfg.MaxSegmentSize {
			ask = c.cfg.MaxSegmentSize
		}
		c.sendMu.Unlock()

		sent, err := c.sendFragment(off, ask)
		if err != nil {
			fmt.Printf("sendFragment error: %v\n", err)
			continue
		}
		if !sent {
			fmt.Println("sendFragment did not send packet (sent=false)")
			continue
		}
	}
}
