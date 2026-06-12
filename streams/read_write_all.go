package streams

import "io"

// ReadAllFrom reads from r into each target in order, stopping at the first error.
func ReadAllFrom(r io.Reader, v ...io.ReaderFrom) (n int64, err error) {
	var m int64
	for _, i := range v {
		m, err = i.ReadFrom(r)
		n += m
		if err != nil {
			return
		}
	}
	return
}

// WriteAllTo writes each source to w in order, stopping at the first error.
func WriteAllTo(w io.Writer, v ...io.WriterTo) (n int64, err error) {
	var m int64
	for _, i := range v {
		m, err = i.WriteTo(w)
		n += m
		if err != nil {
			return
		}
	}
	return
}
