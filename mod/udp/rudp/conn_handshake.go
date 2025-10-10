package rudp

import (
	"context"
	"crypto/rand"
	"fmt"
	"net"
	"time"

	"github.com/cryptopunkscc/astrald/mod/udp"
)

type ConnState int

const (
	StateClosed      ConnState = iota // no connection / after Close()
	StateListen                       // (server only) waiting for SYN
	StateSynSent                      // client sent SYN, waiting for SYN|ACK
	StateSynReceived                  // server got SYN, sent SYN|ACK, waiting for final ACK
	StateEstablished                  // handshake complete, normal data flow
	StateFinSent                      // Close() called, FIN sent, waiting for ACK
	StateFinReceived                  // FIN received, waiting for local Close()
	StateTimeWait                     // (optional) short wait after FIN to absorb retransmits
)

func (c *Conn) StartClientHandshake(ctx context.Context) error {
	return c.startClientHandshakeDirect(ctx)
}

// startClientHandshakeDirect performs the 3-way handshake using direct socket reads.
func (c *Conn) startClientHandshakeDirect(ctx context.Context) error {
	seq, err := randUint32NZ()
	if err != nil {
		return fmt.Errorf("failed to generate initial sequence number: %w", err)
	}
	c.initialSeqNumLocal = seq
	c.connID = seq
	c.setState(StateSynSent)

	if err := c.sendHandshakeControl(FlagSYN, seq, 0); err != nil {
		return err
	}

	buf := make([]byte, 1500)
	deadlineInterval := 300 * time.Millisecond

	for {
		if ctx.Err() != nil {
			return udp.ErrHandshakeTimeout
		}
		_ = c.udpConn.SetReadDeadline(time.Now().Add(deadlineInterval))
		n, addr, err := c.udpConn.ReadFromUDP(buf)

		if err != nil {
			if ne, ok := err.(net.Error); ok && ne.Timeout() {
				continue
			}
			continue
		}
		// Filter by remote IP (ignore port mismatch in case of NAT rebinding; only IP match)
		if c.remoteEndpoint != nil && !addr.IP.Equal(net.IP(c.remoteEndpoint.IP)) {
			continue
		}
		if n < 13 {
			continue
		}
		pkt := &Packet{}
		if err := pkt.Unmarshal(buf[:n]); err != nil {
			continue
		}

		// Expect SYN|ACK
		if pkt.Flags&(FlagSYN|FlagACK) == (FlagSYN|FlagACK) && pkt.Ack == seq+1 && pkt.Seq != 0 {
			c.initialSeqNumRemote = pkt.Seq
			// finalize local cumulative ack/next sequence
			c.sendMu.Lock()
			c.ackedSeqNum = seq + 1
			c.nextSeqNum = seq + 1
			c.sendMu.Unlock()
			// Send final ACK
			if err := c.SendControlPacket(FlagACK, seq+1, c.initialSeqNumRemote+1); err != nil {
				return err
			}
			c.onEstablished()
			// Start recvLoop after establishment for outbound
			go c.recvLoop()
			return nil
		}
		// ignore other control/data until handshake completes
	}
}

func (c *Conn) StartServerHandshake(ctx context.Context, synPkt *Packet) error {
	c.initialSeqNumRemote = synPkt.Seq
	c.connID = synPkt.Seq
	seq, err := randUint32NZ()
	if err != nil {
		return fmt.Errorf("failed to generate initial sequence number: %w", err)
	}
	c.initialSeqNumLocal = seq
	c.setState(StateSynReceived)

	// send SYN|ACK and register for retransmission
	err = c.sendHandshakeControl(FlagSYN|FlagACK, c.initialSeqNumLocal, c.initialSeqNumRemote+1)
	if err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return udp.ErrHandshakeTimeout
		case pkt := <-c.inCh:
			if pkt.Flags&FlagACK != 0 && pkt.Ack == c.initialSeqNumLocal+1 {
				// final ACK received
				c.sendMu.Lock()
				delete(c.unacked, c.initialSeqNumLocal)
				if len(c.unacked) == 0 && c.rtxTimer != nil {
					c.rtxTimer.Stop()
					c.rtxTimer = nil
				}
				c.ackedSeqNum = c.initialSeqNumLocal + 1
				c.nextSeqNum = c.initialSeqNumLocal + 1
				c.sendMu.Unlock()
				c.onEstablished()
				// fused receive loop will now dispatch directly
				return nil
			}
		}
	}
}

// sendHandshakeControl builds, sends and registers a handshake control packet for unified retransmissions
func (c *Conn) sendHandshakeControl(flags uint8, seq, ack uint32) error {
	pkt := &Packet{Seq: seq, Ack: ack, Flags: flags, Len: 0}
	b, err := pkt.Marshal()
	if err != nil {
		return fmt.Errorf("marshal handshake pkt: %w", err)
	}
	if c.udpConn == nil {
		return udp.ErrConnClosed
	}
	if c.remoteEndpoint == nil {
		return fmt.Errorf("remote endpoint nil")
	}

	_, err = c.sendDatagram(b)
	if err != nil {
		return err
	}

	// Register packet and decide if timer start is needed - do this OUTSIDE the lock
	needTimer := false
	c.sendMu.Lock()
	if _, exists := c.unacked[seq]; !exists { // only register first time
		c.unacked[seq] = &Unacked{
			pkt:         pkt,
			sentTime:    time.Now(),
			rtxCount:    0,
			length:      0,
			isHandshake: true,
		}
		if c.rtxTimer == nil {
			needTimer = true
		}
	}
	c.sendMu.Unlock()

	// Start timer AFTER releasing lock to avoid deadlock
	if needTimer {
		c.startRtxTimer()
	}

	return nil
}

// SendControlPacket retained for non-handshake control (e.g., FIN, pure ACK), not tracked
func (c *Conn) SendControlPacket(flags uint8, seq, ack uint32) error {
	pkt := &Packet{Seq: seq, Ack: ack, Flags: flags, Len: 0}
	data, err := pkt.Marshal()
	if err != nil {
		return fmt.Errorf(`SendControlPacket failed to marshal control packet: %w`, err)
	}
	if c.udpConn == nil {
		return udp.ErrConnClosed
	}
	if c.remoteEndpoint == nil {
		return fmt.Errorf("remote endpoint nil")
	}
	_, err = c.sendDatagram(data)
	if err != nil {
		return fmt.Errorf(`SendControlPacket failed to send control packet: %w`, err)
	}
	return nil
}

func randUint32NZ() (uint32, error) {
	var b [4]byte
	for {
		_, err := rand.Read(b[:])
		if err != nil {
			return 0, fmt.Errorf("failed to generate random uint32: %w", err)
		}
		v := uint32(b[0])<<24 | uint32(b[1])<<16 | uint32(b[2])<<8 | uint32(b[3])
		if v != 0 {
			return v, nil
		}
	}
}
