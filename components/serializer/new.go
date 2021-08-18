package serializer

import "io"

func New(rw io.ReadWriteCloser) ReadWriteCloser {
	return &readWriteCloser{
		Closer: rw,
		Reader: NewReader(rw),
		Writer: NewWriter(rw),
	}
}

func NewReader(r io.Reader) Reader {
	return &reader{r}
}

func NewWriter(w io.Writer) Writer {
	return &writer{w}
}
