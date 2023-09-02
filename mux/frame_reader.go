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

func (reader *FrameReader) HandleMux(event Event) {
	switch event := event.(type) {
	case Frame:
		reader.handleFrame(event)

	case Unbind:
		reader.Close()
	}
}

func (reader *FrameReader) handleFrame(frame Frame) {
	if frame.IsEmpty() {
		frame.Mux.Unbind(frame.Port)
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
