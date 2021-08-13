package mux

import (
	"errors"
	"math"
)

// MaxPayload maximum payload size a single mux frame can carry. Frames are 64kb with 4-byte headers.
const MaxPayload = math.MaxUint16 - 4

// MaxStreams number of streams in the multiplexer
const MaxStreams = 65536

// ErrBufferTooBig - buffer size exceeds frame size
var ErrBufferTooBig = errors.New("payload too big")

// ErrBufferTooSmall - buffer too small to read frame
var ErrBufferTooSmall = errors.New("buffer too small")

var ErrInvalidStreamID = errors.New("invalid stream id")

// ErrStreamClosed - stream is closed and cannot be written to or read from
var ErrStreamClosed = errors.New("stream closed")
