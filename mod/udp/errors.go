package udp

import "errors"

var (
	ErrPacketTooShort           = errors.New("packet too short")
	ErrListenerClosed           = errors.New("listener closed")
	ErrConnClosed               = errors.New("connection closed")
	ErrInvalidPayloadLength     = errors.New("invalid payload length")
	ErrClosed                   = errors.New("connection closed")
	ErrZeroMSS                  = errors.New("invalid MSS")
	ErrMalformedPacket          = errors.New("malformed packet")
	ErrHandshakeTimeout         = errors.New("handshake timeout")
	ErrHandshakeReset           = errors.New("handshake reset")
	ErrConnectionNotEstablished = errors.New("connection not established")
)
