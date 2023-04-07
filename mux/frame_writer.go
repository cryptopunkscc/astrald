package mux

import (
	"sync"
)

type FrameWriter struct {
	mu   sync.Mutex
	mux  *FrameMux
	port int
	err  error
}

func (writer *FrameWriter) Port() int {
	return writer.port
}

func NewFrameWriter(mux *FrameMux, port int) *FrameWriter {
	return &FrameWriter{mux: mux, port: port}
}

func (writer *FrameWriter) Write(p []byte) (n int, err error) {
	left := p[:]

	for len(left) > 0 {
		chunkLen := MaxFrameSize
		if chunkLen > len(left) {
			chunkLen = len(left)
		}

		writer.mu.Lock()
		if writer.err == nil {
			err = writer.mux.mux.Write(writer.port, left[0:chunkLen])
		} else {
			err = writer.err
		}
		writer.mu.Unlock()
		if err != nil {
			return
		}

		n += chunkLen
		left = left[chunkLen:]
	}

	return
}

func (writer *FrameWriter) Close() error {
	writer.mu.Lock()
	defer writer.mu.Unlock()

	if writer.err == nil {
		writer.mux.Close(writer.port)
		writer.err = ErrPortClosed
	}

	return nil
}
