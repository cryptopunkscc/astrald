package udp

import (
	"net"
	"time"
)

// recvLoop fuses raw UDP receive and packet dispatch. During handshake (state < Established)
// packets are delivered into inCh for the blocking handshake loops. After establishment,
// packets are dispatched directly to the appropriate handlers without using inCh.
func (c *Conn) recvLoop() {
	const maxPayloadSize = 64 * 1024
	buf := make([]byte, maxPayloadSize)
	for {
		if c.isClosed() {
			return
		}
		if err := c.udpConn.SetReadDeadline(time.Now().Add(1 * time.Second)); err != nil {
			continue
		}
		n, addr, err := c.udpConn.ReadFromUDP(buf)
		if err != nil {
			if ne, ok := err.(net.Error); ok && ne.Timeout() {
				continue
			}
			if c.isClosed() {
				return
			}
			continue
		}
		// address filter
		if addr.String() != c.remoteEndpoint.IP.String() {
			continue
		}
		if n < 13 { // minimum header size
			continue
		}
		packetData := make([]byte, n)
		copy(packetData, buf[:n])
		pkt := &Packet{}
		if err := pkt.Unmarshal(packetData); err != nil {
			continue
		}
		if int(pkt.Len) > maxPayloadSize { // sanity
			continue
		}

		// Handshake phase: deliver via channel for Start*Handshake loops
		if !c.inState(StateEstablished) {
			select {
			case c.inCh <- pkt:
			default: // if channel full (unlikely in handshake), drop
			}
			continue
		}

		// Established: direct dispatch
		if pkt.Flags&FlagACK != 0 {
			c.HandleAckPacket(pkt)
			continue
		}
		if pkt.Flags&(FlagSYN|FlagFIN) != 0 {
			c.HandleControlPacket(pkt)
			continue
		}
		c.handleDataPacket(pkt)
	}
}

func (c *Conn) handleDataPacket(packet *Packet) {
	if packet.Len == 0 { // ignore empty as data
		return
	}
	seq := packet.Seq
	plen := uint32(packet.Len)

	c.recvMu.Lock()
	exp := c.expected
	ackDelay := c.cfg.AckDelay // snapshot once inside lock

	switch {
	case seq < exp: // duplicate / already received
		c.queueAckLocked()
		c.recvMu.Unlock()
		c.triggerAck(ackDelay)
		return

	case seq == exp: // in-order
		if int(packet.Len) > int(c.recvRB.Free()) {
			// No space -> request ACK (window advertisement) and drop
			c.queueAckLocked()
			c.recvMu.Unlock()
			c.triggerAck(ackDelay)
			return
		}
		// Write this packet fully
		if n, _ := c.recvRB.Write(packet.Payload); n != int(packet.Len) {
			// Failed partial write (shouldn't happen with ringbuffer); request ACK and abort
			c.queueAckLocked()
			c.recvMu.Unlock()
			c.triggerAck(ackDelay)
			return
		}
		exp += plen
		c.expected = exp
		// Drain any now-contiguous buffered packets
		for {
			nextPkt, ok := c.recvOO[exp]
			if !ok {
				break
			}
			if int(nextPkt.Len) > int(c.recvRB.Free()) {
				// Break if capacity insufficient; will retry on next in-order arrival
				break
			}
			if n, _ := c.recvRB.Write(nextPkt.Payload); n != int(nextPkt.Len) {
				// On partial write, stop draining to preserve consistency
				break
			}
			delete(c.recvOO, exp)
			exp += uint32(nextPkt.Len)
			c.expected = exp
		}
		c.queueAckLocked()
		c.recvCond.Broadcast()
		c.recvMu.Unlock()

		// mirror expected -> ackedSeqNum for piggyback
		c.sendMu.Lock()
		if c.ackedSeqNum < exp {
			c.ackedSeqNum = exp
		}
		c.sendMu.Unlock()
		c.triggerAck(ackDelay)
		return

	default: // seq > exp (future / out-of-order)
		if int(packet.Len) <= int(c.recvRB.Free()) {
			if _, exists := c.recvOO[seq]; !exists {
				c.recvOO[seq] = packet
			}
		}
		c.queueAckLocked() // request duplicate ACK
		c.recvMu.Unlock()
		c.triggerAck(ackDelay)
		return
	}
}
