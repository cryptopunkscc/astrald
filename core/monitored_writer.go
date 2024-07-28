package core

import (
	"github.com/cryptopunkscc/astrald/id"
	"github.com/cryptopunkscc/astrald/sig"
	"io"
)

type MonitoredWriter struct {
	io.WriteCloser
	io.Reader
	sig.Activity
	identity   id.Identity
	bytes      int
	AfterWrite func(int, error)
	AfterClose func(error)
}

func (w *MonitoredWriter) Identity() id.Identity {
	return w.identity
}

func NewMonitoredWriter(w io.WriteCloser, identity id.Identity) *MonitoredWriter {
	return &MonitoredWriter{
		WriteCloser: w,
		identity:    identity,
	}
}

func (w *MonitoredWriter) Write(p []byte) (n int, err error) {
	defer w.Touch()

	n, err = w.WriteCloser.Write(p)
	w.bytes += n

	if w.AfterWrite != nil {
		w.AfterWrite(n, err)
	}
	return
}

func (w *MonitoredWriter) Close() (err error) {
	defer w.Touch()

	err = w.WriteCloser.Close()
	if w.AfterClose != nil {
		w.AfterClose(err)
	}
	return
}

func (w *MonitoredWriter) Bytes() int {
	return w.bytes
}
