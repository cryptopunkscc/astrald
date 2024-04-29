package mem

import (
	"github.com/cryptopunkscc/astrald/mod/objects"
	"io"
	"sync/atomic"
)

var _ objects.Reader = &Reader{}

type Reader struct {
	bytes  []byte
	offset int64
	closed atomic.Bool
}

func NewMemDataReader(bytes []byte) *Reader {
	return &Reader{bytes: bytes}
}

func (r *Reader) Read(p []byte) (n int, err error) {
	if r.closed.Load() {
		return 0, objects.ErrClosedPipe
	}

	if len(p) == 0 {
		return 0, nil
	}
	if len(r.bytes)-int(r.offset) == 0 {
		return 0, io.EOF
	}

	var end = min(int(r.offset)+len(p), len(r.bytes))
	n = copy(p, r.bytes[r.offset:end])

	r.offset += int64(n)

	return n, nil
}

func (r *Reader) Seek(offset int64, whence int) (int64, error) {
	if r.closed.Load() {
		return 0, objects.ErrClosedPipe
	}

	var l = int64(len(r.bytes))
	var o = r.offset

	switch whence {
	case io.SeekStart:
		o = offset
	case io.SeekCurrent:
		o = o + offset
	case io.SeekEnd:
		o = l + offset
	}

	r.offset = max(min(o, l), 0)

	return r.offset, nil
}

func (r *Reader) Close() error {
	r.closed.Store(true)
	return nil
}

func (r *Reader) Info() *objects.ReaderInfo {
	return &objects.ReaderInfo{Name: "memory"}
}
