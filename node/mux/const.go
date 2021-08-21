package mux

import (
	"errors"
	"math"
)

// MaxFrameSize is the maximum total size of a mux frame including the header
const MaxFrameSize = math.MaxUint16

// headerLen is the length of the frame header (uint16 for streamID and uint16 for payload size)
const headerLen = 4

// MaxPayload maximum payload size a single mux frame can carry
const MaxPayload = MaxFrameSize - headerLen

// MaxStreams number of streams in the multiplexer
const MaxStreams = 1 << 16

// ErrBufferTooBig - buffer size exceeds frame size
var ErrBufferTooBig = errors.New("payload too big")

// ErrBufferTooSmall - buffer too small to read frame
var ErrBufferTooSmall = errors.New("buffer too small")

// ErrInvalidStreamID - stream ID is outside the valid range
var ErrInvalidStreamID = errors.New("invalid stream id")

// ErrStreamClosed - stream is closed and cannot be written to or read from
var ErrStreamClosed = errors.New("stream closed")
