package mux

import (
	"errors"
	"fmt"
)

// ErrBufferTooBig - buffer size exceeds frame size
var ErrBufferTooBig = errors.New("payload too big")

// ErrBufferTooSmall - buffer too small to read frame
var ErrBufferTooSmall = errors.New("buffer too small")

// ErrInvalidStreamID - stream ID is outside the valid range
var ErrInvalidStreamID = errors.New("invalid stream id")

// ErrStreamClosed - stream is closed and cannot be written to or read from
var ErrStreamClosed = errors.New("stream closed")

// ErrDemuxClosed - demux session is closed
var ErrDemuxClosed = errors.New("demux closed")

// ErrStreamLimitReached - all inputs streams are already in use
var ErrStreamLimitReached = errors.New("stream limit reached")

// ErrFrameOnClosedStream - received a frame on a closed input stream
type ErrFrameOnClosedStream struct {
	id int
}

func (err ErrFrameOnClosedStream) Error() string {
	return fmt.Sprintf("read frame on a closed stream (%d)", err.id)
}

var errControlStreamClosed = errors.New("control stream closed")
