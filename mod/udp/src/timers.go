package udp

import (
	"time"
)

// armRTO starts or resets the retransmission timer
func (c *Conn) armRTO(d time.Duration) {
	c.rtoMu.Lock()
	defer c.rtoMu.Unlock()

	if c.rtoTimer != nil {
		c.rtoTimer.Stop()
	}
	c.rtoTimer = time.AfterFunc(d, c.handleRTO)
}

// stopRTO safely stops the retransmission timer
func (c *Conn) stopRTO() {
	c.rtoMu.Lock()
	defer c.rtoMu.Unlock()

	if c.rtoTimer != nil {
		c.rtoTimer.Stop()
		c.rtoTimer = nil
	}
}

// armAckDelay schedules a pure ACK to be sent soon
func (c *Conn) armAckDelay() {
	c.rtoMu.Lock()
	defer c.rtoMu.Unlock()

	if c.ackTimer == nil {
		c.ackTimer = time.AfterFunc(c.cfg.AckDelay, c.sendPureACK)
	} else {
		c.ackTimer.Reset(c.cfg.AckDelay)
	}
}

// armAckDelayTimerLocked initializes the ACK delay timer (called during initialization)
// Note: Unlike armAckDelay, this is called from NewConn when mutex is already held
func (c *Conn) armAckDelayTimerLocked() {
	c.ackTimer = time.AfterFunc(c.cfg.AckDelay, c.sendPureACK)
}
