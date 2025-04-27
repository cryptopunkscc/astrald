package mem

import (
	"bytes"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/object"
	"sync/atomic"
)

var _ objects.Writer = &Writer{}

type Writer struct {
	*Repository
	buf    *bytes.Buffer
	closed atomic.Bool
}

func NewWriter(memStore *Repository) *Writer {
	return &Writer{
		Repository: memStore,
		buf:        &bytes.Buffer{},
	}
}

func (w *Writer) Write(p []byte) (n int, err error) {
	if w.closed.Load() {
		return 0, objects.ErrClosedPipe
	}

	if int64(len(p)) > w.free() {
		return 0, objects.ErrNoSpaceLeft
	}
	n, err = w.buf.Write(p)
	w.used.Add(int64(n))
	return n, err
}

func (w *Writer) Commit() (*object.ID, error) {
	if !w.closed.CompareAndSwap(false, true) {
		return nil, objects.ErrClosedPipe
	}

	var buf = w.buf.Bytes()
	var objectID, _ = object.Resolve(bytes.NewReader(buf))

	w.objects.Set(objectID.String(), buf)

	return objectID, nil
}

func (w *Writer) Discard() error {
	if w.closed.CompareAndSwap(false, true) {
		w.used.Add(int64(-w.buf.Len())) // free up space
		w.buf = nil
	}
	return nil
}
