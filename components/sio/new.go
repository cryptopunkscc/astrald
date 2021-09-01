package sio

import "io"

func New(rw io.ReadWriteCloser) ReadWriteCloser {
	return &readWriteCloser{
		Closer: rw,
		Reader: NewReader(rw),
		Writer: NewWriter(rw),
	}
}

func NewReader(r io.Reader) Reader {
	return &reader{Reader: r}
}

func NewReadCloser(r io.ReadCloser) ReadCloser {
	return &reader{r, r}
}

func NewWriter(w io.Writer) Writer {
	return &writer{Writer: w}
}

func NewWritCloser(w io.WriteCloser) WriteCloser {
	return &writer{w, w}
}
