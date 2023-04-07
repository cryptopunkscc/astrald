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

func (reader *FrameReader) HandleFrame(frame Frame) error {
	if frame.EOF() {
		reader.w.Close()
		return io.EOF
	}

	_, err := reader.w.Write(frame.Data)
	return err
}

func (reader *FrameReader) Read(p []byte) (int, error) {
	return reader.r.Read(p)
}
