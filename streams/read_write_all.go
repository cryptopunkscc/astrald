package streams

import "io"

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
