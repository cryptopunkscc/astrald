package net

import "io"

type PipeReader struct {
	io.Reader
	pipeWriter *PipeWriter
}

type PipeWriter struct {
	io.WriteCloser
	source any
}

type Sourcer interface {
	Source() any
	SetSource(source any)
}

var _ Sourcer = &PipeReader{}
var _ Sourcer = &PipeWriter{}

func Pipe() (*PipeReader, *PipeWriter) {
	rr, ww := io.Pipe()

	w := newPipeWriter(ww)
	r := newPipeReader(rr, w)

	return r, w
}

func newPipeReader(r io.Reader, w *PipeWriter) *PipeReader {
	return &PipeReader{Reader: r, pipeWriter: w}
}

func newPipeWriter(w io.WriteCloser) *PipeWriter {
	return &PipeWriter{WriteCloser: w}
}

func (p *PipeReader) Source() any {
	return p.pipeWriter.Source()
}

func (p *PipeReader) SetSource(source any) {
	p.pipeWriter.SetSource(source)
}

func (w *PipeWriter) Source() any {
	return w.source
}

func (w *PipeWriter) SetSource(source any) {
	w.source = source
}
