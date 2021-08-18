package serializer

import "io"

func New(rw io.ReadWriteCloser) *ReadWriteCloser {
	return &ReadWriteCloser{
		Closer: rw,
		Reader: NewReader(rw),
		Writer: NewWriter(rw),
	}
}

func NewReader(reader io.Reader) *Reader {
	return &Reader{reader}
}

func NewWriter(writer io.Writer) *Writer {
	return &Writer{writer}
}