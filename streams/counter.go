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

type WriteCounter struct {
	w io.Writer
	n int64
}

func NewWriteCounter(w io.Writer) *WriteCounter {
	if w == nil {
		w = NilWriter{}
	}
	return &WriteCounter{w: w}
}

func (w *WriteCounter) Write(p []byte) (n int, err error) {
	n, err = w.w.Write(p)
	w.n += int64(n)
	return
}

func (w *WriteCounter) Total() int64 {
	return w.n
}
