package mux

import (
	"io"
)

type FrameReader struct {
	r *io.PipeReader
	w *io.PipeWriter
}

func NewFrameReader() *FrameReader {
	r, w := io.Pipe()
	return &FrameReader{
		r: r,
		w: w,
	}
}

func (reader *FrameReader) HandleFrame(frame Frame) {
	if frame.IsEmpty() {
		reader.w.Close()
		return
	}

	reader.w.Write(frame.Data)
	return
}

func (reader *FrameReader) Read(p []byte) (int, error) {
	return reader.r.Read(p)
}

// Close closes the writer side of the pipe
func (reader *FrameReader) Close() error {
	return reader.w.Close()
}
