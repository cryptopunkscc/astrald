package mux

import (
	"sync"
)

type OutputStream struct {
	id        int
	mux       *Mux
	closeOnce sync.Once
	err       error
}

func NewOutputStream(mux *Mux, streamID int) *OutputStream {
	return &OutputStream{
		id:  streamID,
		mux: mux,
	}
}

// Write writes a byte buffer to the connectiontion
func (stream *OutputStream) Write(data []byte) (n int, err error) {
	if stream.err != nil {
		return 0, stream.err
	}

	left := data[:]

	for len(left) > 0 {
		chunkLen := MaxPayload
		if chunkLen > len(left) {
			chunkLen = len(left)
		}

		err = stream.mux.Write(stream.id, left[0:chunkLen])
		if err != nil {
			return
		}

		n += chunkLen
		left = left[chunkLen:]
	}

	return
}

// Close closes the stream
func (stream *OutputStream) Close() (err error) {
	err = ErrStreamClosed

	stream.closeOnce.Do(func() {
		err = nil

		stream.err = stream.mux.Write(stream.id, nil)
		if stream.err == nil {
			stream.err = ErrStreamClosed
		} else {
			err = stream.err
		}
	})

	return
}

func (stream OutputStream) ID() int {
	return stream.id
}
