package fs

import (
	"bytes"
	"github.com/cryptopunkscc/astrald/data"
	"github.com/cryptopunkscc/astrald/mod/storage"
	"sync/atomic"
)

var _ storage.DataWriter = &MemDataWriter{}

type MemDataWriter struct {
	*MemStore
	buf    *bytes.Buffer
	closed atomic.Bool
}

func NewMemDataWriter(memStore *MemStore) *MemDataWriter {
	return &MemDataWriter{
		MemStore: memStore,
		buf:      &bytes.Buffer{},
	}
}

func (w *MemDataWriter) Write(p []byte) (n int, err error) {
	if w.closed.Load() {
		return 0, storage.ErrClosedPipe
	}
	if int64(len(p)) > w.Free() {
		return 0, storage.ErrNoSpaceLeft
	}
	n, err = w.buf.Write(p)
	w.used.Add(int64(n))
	return n, err
}

func (w *MemDataWriter) Commit() (data.ID, error) {
	if !w.closed.CompareAndSwap(false, true) {
		return data.ID{}, storage.ErrClosedPipe
	}

	var buf = w.buf.Bytes()
	var dataID = data.Resolve(buf)

	if _, ok := w.objects.Set(dataID.String(), buf); ok {
		w.events.Emit(storage.EventDataCommitted{DataID: dataID})
	}

	return dataID, nil
}

func (w *MemDataWriter) Discard() error {
	if w.closed.CompareAndSwap(false, true) {
		w.used.Add(int64(-w.buf.Len())) // free up space
		w.buf = nil
	}
	return nil
}
