package streams

import "io"

type ReadCounter struct {
	r io.Reader
	n int64
}

func NewReadCounter(r io.Reader) *ReadCounter {
	return &ReadCounter{r: r}
}

func (r *ReadCounter) Read(p []byte) (n int, err error) {
	n, err = r.r.Read(p)
	r.n += int64(n)
	return
}

func (r *ReadCounter) Total() int64 {
	return r.n
}
