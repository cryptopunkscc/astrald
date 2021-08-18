package adapter

import "io"

type reader struct {
	io.ReadCloser
	size int64
}

func (r reader) Size() (int64, error) {
	return r.size, nil
}