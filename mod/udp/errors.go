package udp

import "errors"

var (
	ErrListenerClosed              = errors.New("listener closed")
	ErrRetransmissionLimitExceeded = errors.New(
		"retransmissions limit exceeded")

	// ErrDataLost is emitted on close if there was still buffered or unacked data.
	ErrDataLost                 = errors.New("unsent data lost on close")
	ErrPacketTooShort           = errors.New("packet too short")
	ErrConnClosed               = errors.New("connection closed")
	ErrInvalidPayloadLength     = errors.New("invalid payload length")
	ErrZeroMSS                  = errors.New("invalid MSS")
	ErrMalformedPacket          = errors.New("malformed packet")
	ErrHandshakeTimeout         = errors.New("handshake timeout")
	ErrConnectionNotEstablished = errors.New("connection not established")
)
