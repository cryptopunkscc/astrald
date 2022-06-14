package streams

import "io"

var _ io.ReadCloser = &LimitedReader{}

type LimitedReader struct {
	io.ReadCloser
	Limit uint64
	n     uint64
}

func (l *LimitedReader) Read(p []byte) (n int, err error) {
	if l.n == l.Limit {
		return 0, io.EOF
	}

	plen := uint64(len(p))

	if l.n+plen > l.Limit {
		plen = l.Limit - l.n
	}

	n, err = l.ReadCloser.Read(p[:plen])
	l.n += uint64(n)

	return
}

func (l *LimitedReader) Close() error {
	return l.ReadCloser.Close()
}
