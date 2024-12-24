package mem

import (
	"bytes"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/object"
	"sync/atomic"
)

var _ objects.Writer = &Writer{}

type Writer struct {
	*Store
	buf    *bytes.Buffer
	closed atomic.Bool
}

func NewMemDataWriter(memStore *Store) *Writer {
	return &Writer{
		Store: memStore,
		buf:   &bytes.Buffer{},
	}
}

func (w *Writer) Write(p []byte) (n int, err error) {
	if w.closed.Load() {
		return 0, objects.ErrClosedPipe
	}
	if int64(len(p)) > w.Free() {
		return 0, objects.ErrNoSpaceLeft
	}
	n, err = w.buf.Write(p)
	w.used.Add(int64(n))
	return n, err
}

func (w *Writer) Commit() (object.ID, error) {
	if !w.closed.CompareAndSwap(false, true) {
		return object.ID{}, objects.ErrClosedPipe
	}

	var buf = w.buf.Bytes()
	var objectID, _ = object.Resolve(bytes.NewReader(buf))

	if _, ok := w.objects.Set(objectID.String(), buf); ok {
		w.mod.Receive(&objects.EventCommitted{ObjectID: objectID}, nil)
	}

	return objectID, nil
}

func (w *Writer) Discard() error {
	if w.closed.CompareAndSwap(false, true) {
		w.used.Add(int64(-w.buf.Len())) // free up space
		w.buf = nil
	}
	return nil
}
