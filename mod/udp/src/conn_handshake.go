package udp

import (
	"context"
	"crypto/rand"
	"fmt"
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
	seq, err := randUint32NZ()
	if err != nil {
		return fmt.Errorf("failed to generate initial sequence number: %w", err)
	}
	c.initialSeqNumLocal = seq
	c.connID = c.initialSeqNumLocal
	c.setState(StateSynSent)

	// build + send initial SYN and register in unacked for unified retransmission
	if err := c.sendHandshakeControl(FlagSYN, c.initialSeqNumLocal, 0); err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return udp.ErrHandshakeTimeout
		case pkt := <-c.inCh:
			if pkt.Flags&(FlagSYN|FlagACK) == (FlagSYN|FlagACK) && pkt.Ack == c.initialSeqNumLocal+1 && pkt.Seq != 0 {
				// got valid SYN|ACK
				c.initialSeqNumRemote = pkt.Seq
				// remove our SYN from unacked and set sequence bases
				c.sendMu.Lock()
				delete(c.unacked, c.initialSeqNumLocal)
				if len(c.unacked) == 0 && c.rtxTimer != nil {
					c.rtxTimer.Stop()
					c.rtxTimer = nil
				}
				c.ackedSeqNum = c.initialSeqNumLocal + 1
				c.sendBase = c.initialSeqNumLocal + 1
				c.nextSeqNum = c.initialSeqNumLocal + 1
				c.sendMu.Unlock()

				// send final ACK (not tracked for retransmission)
				if err := c.SendControlPacket(FlagACK, c.initialSeqNumLocal+1, c.initialSeqNumRemote+1); err != nil {
					return err
				}
				// Transition to established (state change + sender loop) via helper
				c.onEstablished()
				// fused receive loop will now dispatch directly
				return nil
			}
		}
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
	if err := c.sendHandshakeControl(FlagSYN|FlagACK, c.initialSeqNumLocal, c.initialSeqNumRemote+1); err != nil {
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
				c.sendBase = c.initialSeqNumLocal + 1
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
	// Use WriteToUDP to support server-side unconnected sockets and client connected sockets uniformly.
	if c.remoteEndpoint == nil {
		return fmt.Errorf("remote endpoint nil")
	}
	if _, err := c.udpConn.WriteToUDP(b, c.remoteEndpoint.UDPAddr()); err != nil {
		return err
	}
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
			c.startRtxTimer()
		}
	}
	c.sendMu.Unlock()
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
	_, err = c.udpConn.WriteToUDP(data, c.remoteEndpoint.UDPAddr())
	if err != nil {
		return fmt.Errorf(`SendControlPacket failed to send control packet: %w`, err)
	}
	return err
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
