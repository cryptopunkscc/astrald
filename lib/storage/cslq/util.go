package storage

import "io"

type rwClose struct {
	io.ReadWriter
	close func() error
}

func (rwc rwClose) Close() error {
	return rwc.close()
}
