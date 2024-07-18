package core

import (
	"github.com/cryptopunkscc/astrald/id"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/sig"
)

var _ astral.SecureWriteCloser = &MonitoredWriter{}
var _ astral.OutputGetter = &MonitoredWriter{}

type MonitoredWriter struct {
	*astral.SourceField
	*astral.OutputField
	sig.Activity
	bytes      int
	AfterWrite func(int, error)
	AfterClose func(error)
}

func (w *MonitoredWriter) Identity() id.Identity {
	return w.Output().Identity()
}

func NewMonitoredWriter(w astral.SecureWriteCloser) *MonitoredWriter {
	m := &MonitoredWriter{
		SourceField: astral.NewSourceField(nil),
	}
	m.OutputField = astral.NewOutputField(m, w)

	return m
}

func (w *MonitoredWriter) Write(p []byte) (n int, err error) {
	defer w.Touch()

	n, err = w.Output().Write(p)
	w.bytes += n

	if w.AfterWrite != nil {
		w.AfterWrite(n, err)
	}
	return
}

func (w *MonitoredWriter) Close() (err error) {
	defer w.Touch()

	err = w.Output().Close()
	if w.AfterClose != nil {
		w.AfterClose(err)
	}
	return
}

func (w *MonitoredWriter) Bytes() int {
	return w.bytes
}
