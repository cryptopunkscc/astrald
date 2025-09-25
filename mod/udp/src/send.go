package udp

import (
	"context"
	"time"

	"github.com/cryptopunkscc/astrald/mod/udp"
)

// Write implements batched segmentation and send.
func (c *Conn) Write(p []byte) (int, error) {
	if c.closed.Load() {
		return 0, udp.ErrClosed
	}
	written := 0

	for written < len(p) {
		chunk := p[written:]

		c.sendMu.Lock()
		for c.sendQ.Len() >= c.cfg.SendBufBytes && !c.closed.Load() {
			c.sendCond.Wait()
		}
		if c.closed.Load() {
			c.sendMu.Unlock()
			return written, udp.ErrClosed
		}
		space := c.cfg.SendBufBytes - c.sendQ.Len()
		if space > 0 {
			toCopy := len(chunk)
			if toCopy > space {
				toCopy = space
			}
			_, err := c.sendQ.Write(chunk[:toCopy])
			if err != nil {
				c.sendMu.Unlock()
				return written, err
			}
			written += toCopy
		}
		c.sendMu.Unlock()

		if err := c.flushSendQueue(); err != nil {
			return written, err
		}
	}

	return written, nil
}

// flushSendQueue cuts ≤ MSS segments from sendQ while within WindowBytes,
// builds packets (piggybacking ACK), records unacked, and writes them back-to-back.
func (c *Conn) flushSendQueue() error {
	var bufs [][]byte
	var ack uint32

	c.sendMu.Lock()
	c.rcvMu.Lock()
	ack = c.rcvNext
	c.rcvMu.Unlock()

	win := c.cfg.WindowBytes - c.bytesInFlight
	for win > 0 && c.sendQ.Len() > 0 {
		segLen := c.mss
		if segLen > c.sendQ.Len() {
			segLen = c.sendQ.Len()
		}
		if segLen > win {
			segLen = win
		}
		if segLen <= 0 {
			break
		}

		payload := c.sendQ.Next(segLen)
		seq := c.nextSeq

		pkt := Packet{
			Seq:     seq,
			Ack:     ack,
			Flags:   FlagACK,
			Win:     0,
			Len:     uint16(segLen),
			Payload: payload,
		}
		raw, err := pkt.Marshal()
		if err != nil {
			c.sendMu.Unlock()
			return err
		}
		bufs = append(bufs, raw)

		meta := segMeta{
			data:     append([]byte(nil), payload...),
			sentAt:   time.Now(),
			retries:  0,
			seqStart: seq,
			length:   segLen,
		}
		c.unacked[seq] = meta
		c.order = append(c.order, seq)

		c.nextSeq += uint32(segLen)
		c.bytesInFlight += segLen
		win -= segLen
	}
	c.sendMu.Unlock()

	if len(bufs) == 0 {
		return nil
	}

	c.writeMu.Lock()
	var writeErr error
	for _, b := range bufs {
		if _, writeErr = c.udpConn.Write(b); writeErr != nil {
			break
		}
	}
	c.writeMu.Unlock()
	if writeErr != nil {
		return writeErr
	}

	c.startRTOIfNeededLocked()
	return nil
}

// startRTOIfNeededLocked arms the RTO timer when unacked is non-empty and no timer running.
func (c *Conn) startRTOIfNeededLocked() {
	c.sendMu.Lock()
	need := len(c.unacked) > 0
	c.sendMu.Unlock()
	if !need {
		return
	}

	c.rtoMu.Lock()
	defer c.rtoMu.Unlock()
	if c.rtoTimer != nil {
		return
	}
	d := c.rto
	if d <= 0 {
		d = c.cfg.RTO
	}
	c.rtoTimer = time.AfterFunc(d, c.onRTOTimeout)
}

// onRTOTimeout retransmits the earliest unacked segment with backoff.
func (c *Conn) onRTOTimeout() {
	var seq uint32
	var meta segMeta
	var ok bool

	c.sendMu.Lock()
	if len(c.order) == 0 {
		c.sendMu.Unlock()
		c.stopRTO()
		return
	}
	seq = c.order[0]
	meta, ok = c.unacked[seq]
	if !ok {
		c.order = c.order[1:]
		c.sendMu.Unlock()
		c.startRTOIfNeededLocked()
		return
	}

	c.rcvMu.Lock()
	ack := c.rcvNext
	c.rcvMu.Unlock()

	pkt := Packet{
		Seq:     meta.seqStart,
		Ack:     ack,
		Flags:   FlagACK,
		Win:     0,
		Len:     uint16(meta.length),
		Payload: meta.data,
	}
	raw, err := pkt.Marshal()
	c.sendMu.Unlock()
	if err != nil {
		c.closeWithError(err)
		return
	}

	c.writeMu.Lock()
	_, werr := c.udpConn.Write(raw)
	c.writeMu.Unlock()
	if werr != nil {
		c.closeWithError(werr)
		return
	}

	c.sendMu.Lock()
	meta.retries++
	meta.sentAt = time.Now()
	c.unacked[seq] = meta
	c.rto *= 2
	if c.rto > c.cfg.RTOMax {
		c.rto = c.cfg.RTOMax
	}
	overLimit := meta.retries > c.cfg.RetryLimit
	c.sendMu.Unlock()

	if overLimit {
		c.closeWithError(context.DeadlineExceeded)
		return
	}

	c.rtoMu.Lock()
	if c.rtoTimer != nil {
		c.rtoTimer.Reset(c.rto)
	}
	c.rtoMu.Unlock()
}

// advanceAck removes fully-acked segments up to 'ack' and manages timers/backpressure.
func (c *Conn) advanceAck(ack uint32) {
	c.sendMu.Lock()
	changed := false
	for len(c.order) > 0 {
		seq := c.order[0]
		meta := c.unacked[seq]
		end := seq + uint32(meta.length)
		// if segment end <= ack, it is fully acked
		if seqLT(end, ack) || end == ack {
			delete(c.unacked, seq)
			c.order = c.order[1:]
			c.bytesInFlight -= meta.length
			changed = true
		} else {
			break
		}
	}
	empty := len(c.unacked) == 0
	c.sendMu.Unlock()

	if changed {
		// let writers proceed if sendQ was full
		c.sendCond.Signal()
	}
	if empty {
		c.stopRTO()
	}
}
