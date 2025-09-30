package udp

import "errors"

var (
	ErrRetransmissionLimitExceeded = errors.New(
		"retransmissions limit exceeded")
	ErrPacketTooShort           = errors.New("packet too short")
	ErrConnClosed               = errors.New("connection closed")
	ErrInvalidPayloadLength     = errors.New("invalid payload length")
	ErrZeroMSS                  = errors.New("invalid MSS")
	ErrMalformedPacket          = errors.New("malformed packet")
	ErrHandshakeTimeout         = errors.New("handshake timeout")
	ErrConnectionNotEstablished = errors.New("connection not established")
)
