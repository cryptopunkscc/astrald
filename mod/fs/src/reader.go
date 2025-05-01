package fs

import (
	"github.com/cryptopunkscc/astrald/mod/objects"
	"io"
	"os"
)

var _ objects.Reader = &Reader{}

type Reader struct {
	io.ReadSeekCloser
	name  string
	limit int64
}

// NewReader returns a new file reader with a limit on the amount of bytes that can be read. -1 means no limit.
func NewReader(f *os.File, name string, limit int64) *Reader {
	return &Reader{
		ReadSeekCloser: f,
		name:           name,
		limit:          limit,
	}
}

func (r *Reader) Read(p []byte) (n int, err error) {
	switch {
	case r.limit < 0:
		return r.ReadSeekCloser.Read(p)
	case r.limit == 0:
		return 0, io.EOF
	}

	l := min(len(p), int(r.limit))
	n, err = r.ReadSeekCloser.Read(p[:l])
	r.limit -= int64(n)

	return n, err
}

func (r *Reader) Seek(offset int64, whence int) (int64, error) {
	return r.ReadSeekCloser.Seek(offset, whence)
}

func (r *Reader) Close() error {
	return r.ReadSeekCloser.Close()
}
