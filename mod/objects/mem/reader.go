package mem

import (
	"bytes"
	"errors"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"os"
	"sync/atomic"
)

var _ objects.Reader = &Reader{}

type Reader struct {
	r      *bytes.Reader
	bytes  []byte
	closed atomic.Bool
}

func NewReader(buf []byte) *Reader {
	return &Reader{
		r:     bytes.NewReader(buf),
		bytes: buf,
	}
}

func (r *Reader) Read(p []byte) (n int, err error) {
	if r.closed.Load() {
		return 0, os.ErrClosed
	}

	return r.r.Read(p)
}

func (r *Reader) Seek(offset int64, whence int) (int64, error) {
	return 0, errors.ErrUnsupported
}

func (r *Reader) Close() error {
	r.closed.Store(true)
	return nil
}
