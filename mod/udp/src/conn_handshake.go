package udp

import (
	"context"
	"fmt"

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

func (c *Conn) startClientHandshake(ctx context.Context) error {
	c.initialSeqNumLocal = randUint32NZ()
	c.connID = c.initialSeqNumLocal
	c.setState(StateSynSent)

	err := c.sendControl(FlagSYN, c.initialSeqNumLocal, 0)
	if err != nil {
		return err
	}

	for {
		select {
		case <-ctx.Done():
			return udp.ErrHandshakeTimeout
		case pkt := <-c.inCh:
			if pkt.Flags&(FlagSYN|FlagACK) == (FlagSYN|FlagACK) && pkt.Ack == c.initialSeqNumLocal+1 && pkt.Seq != 0 {
				c.initialSeqNumRemote = pkt.Seq

				err := c.sendControl(FlagACK, c.initialSeqNumLocal+1, c.initialSeqNumRemote+1)
				if err != nil {
					return err
				}
				c.setState(StateEstablished)
				go c.InboundPacketHandler()
				return nil
			}
		}
	}
}

func (c *Conn) startServerHandshake(ctx context.Context, synPkt *Packet) error {
	c.initialSeqNumRemote = synPkt.Seq
	c.connID = synPkt.Seq
	c.initialSeqNumLocal = randUint32NZ()
	c.setState(StateSynReceived)

	if err := c.sendControl(FlagSYN|FlagACK, c.initialSeqNumLocal, c.initialSeqNumRemote+1); err != nil {
		return err
	}
	for {
		select {
		case <-ctx.Done():
			return udp.ErrHandshakeTimeout
		case pkt := <-c.inCh:
			if pkt.Flags&FlagACK != 0 && pkt.Ack == c.initialSeqNumLocal+1 {
				c.setState(StateEstablished)
				go c.InboundPacketHandler()
				return nil
			}
		}
	}
}

func (c *Conn) sendControl(flags uint8, seq, ack uint32) error {
	pkt := &Packet{
		Seq:   seq,
		Ack:   ack,
		Flags: flags,
		Len:   0,
	}

	data, err := pkt.Marshal()
	if err != nil {
		return fmt.Errorf(`sendControl failed to marshal control packet: %w`, err)
	}

	if c.udpConn == nil {
		return udp.ErrConnClosed
	}

	_, err = c.udpConn.Write(data)
	if err != nil {
		return fmt.Errorf(`sendControl failed to send control packet: %w`, err)
	}

	return err
}

func (c *Conn) handleInbound(pkt *Packet) {
	// ...process handshake/control/data packets...
}

// TODO: implement proper random non-zero uint32 generator
func randUint32NZ() uint32 {
	// ...generate non-zero random uint32...
	return 1 // stub
}

// notifyInbound is a channel for inbound packets
// ...in Conn struct...
// notifyInbound chan *Packet
