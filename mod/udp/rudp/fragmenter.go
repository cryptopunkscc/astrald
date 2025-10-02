package rudp

import "bytes"

// Fragmenter turns buffered bytes into wire packets and reproduces the exact
// same boundaries for retransmission.
type Fragmenter interface {
	// MakeNew decides payload size and builds a new Packet at nextSeq.
	// 'allowed' is the sender's remaining window in bytes.
	// Returns (packet, payloadLen, ok). ok=false if it chooses not to send (e.g., Nagle).
	MakeNew(nextSeq uint32, allowed int, buf *bytes.Buffer) (*Packet, int, bool)
}

// BasicFragmenter is a simple implementation of Fragmenter that splits a
// buffer into Packets of at most MSS size.
type BasicFragmenter struct {
	MSS int
}

// NewBasicFragmenter creates a new BasicFragmenter with the given maximum
// segment size (MSS).
func NewBasicFragmenter(mss int) *BasicFragmenter {
	return &BasicFragmenter{MSS: mss}
}

// MakeNew implements the Fragmenter interface for BasicFragmenter.
func (f *BasicFragmenter) MakeNew(nextSeq uint32, allowed int,
	buf *bytes.Buffer) (*Packet, int, bool) {
	if f.MSS <= 0 {
		return nil, 0, false
	}
	if buf.Len() == 0 {
		return nil, 0, false
	}
	if allowed <= 0 {
		return nil, 0, false
	}

	maxLen := f.MSS
	if allowed < maxLen {
		maxLen = allowed
	}
	if buf.Len() < maxLen {
		maxLen = buf.Len()
	}

	payload := buf.Bytes()[:maxLen]
	packet := &Packet{
		Seq:     nextSeq,
		Len:     uint16(len(payload)),
		Payload: payload,
		Flags:   FlagACK, // Data packets should have ACK flag set
	}
	return packet, len(payload), true
}
