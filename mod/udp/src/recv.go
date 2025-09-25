package udp

import (
	"net"
	"time"
)

// recvLoop parses incoming datagrams, processes ACKs and data, and coalesces ACKs back.
func (c *Conn) recvLoop() {
	defer c.wg.Done()

	buf := make([]byte, 64*1024)

	for {
		_ = c.udpConn.SetReadDeadline(time.Now().Add(500 * time.Millisecond))
		n, _, err := c.udpConn.ReadFromUDP(buf)
		if err != nil {
			if ne, ok := err.(net.Error); ok && ne.Timeout() {
				if c.closed.Load() || c.ctx.Err() != nil {
					return
				}
				continue
			}
			if !c.closed.Load() {
				c.closeWithError(err)
			}
			return
		}

		pkt, uerr := UnmarshalPacket(buf[:n])
		if uerr != nil {
			// drop malformed
			continue
		}

		if pkt.Ack != 0 || (pkt.Flags&FlagACK) != 0 {
			c.advanceAck(pkt.Ack)
		}

		if pkt.Len > 0 {
			if err := c.handleData(pkt); err != nil {
				c.closeWithError(err)
				return
			}
		}
	}
}

// handleData commits in-order payload to appBuf, buffers out-of-order,
// and schedules a delayed pure ACK for the burst.
func (c *Conn) handleData(pkt *Packet) error {
	ackNeeded := false

	c.rcvMu.Lock()
	switch {
	case pkt.Seq == c.rcvNext:
		payload := append([]byte(nil), pkt.Payload...)
		c.rcvMu.Unlock()

		// Do not hold rcvMu while blocking
		if _, err := c.appBuf.WriteAll(payload); err != nil {
			return err
		}

		c.rcvMu.Lock()
		c.rcvNext += uint32(len(payload))
		ackNeeded = true

		// drain contiguous out-of-order
		for {
			next, ok := c.ooo[c.rcvNext]
			if !ok {
				break
			}
			delete(c.ooo, c.rcvNext)
			data := append([]byte(nil), next...)
			c.rcvMu.Unlock()
			if _, err := c.appBuf.WriteAll(data); err != nil {
				return err
			}
			c.rcvMu.Lock()
			c.rcvNext += uint32(len(data))
		}
		c.rcvMu.Unlock()

	case seqLT(c.rcvNext, pkt.Seq):
		// out-of-order: drop if would exceed RecvBuf cap
		if pkt.Len <= uint16(c.cfg.RecvBufBytes) {
			c.ooo[pkt.Seq] = append([]byte(nil), pkt.Payload...)
			ackNeeded = true // peer will infer gap
		}
		c.rcvMu.Unlock()

	default:
		// duplicate; ignore
		c.rcvMu.Unlock()
	}

	if ackNeeded {
		c.armAckDelay()
	}
	return nil
}

// sendPureACK emits a standalone cumulative ACK for rcvNext.
func (c *Conn) sendPureACK() error {
	// snapshot current cumulative ack
	c.rcvMu.Lock()
	ack := c.rcvNext
	c.rcvMu.Unlock()

	pkt := Packet{
		Seq:   c.nextSeq, // sender ignores Seq on pure ACK
		Ack:   ack,
		Flags: FlagACK,
		Win:   0,
		Len:   0,
	}
	raw, err := pkt.Marshal()
	if err != nil {
		return err
	}

	c.writeMu.Lock()
	_, werr := c.udpConn.Write(raw)
	c.writeMu.Unlock()
	return werr
}
