package node

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/net"
	"github.com/cryptopunkscc/astrald/sig"
)

var _ net.SecureWriteCloser = &MonitoredWriter{}
var _ net.OutputGetter = &MonitoredWriter{}

type MonitoredWriter struct {
	*net.SourceField
	*net.OutputField
	sig.Activity
	bytes      int
	AfterWrite func(int, error)
	AfterClose func(error)
}

func (w *MonitoredWriter) Identity() id.Identity {
	return w.Output().Identity()
}

func NewMonitoredWriter(w net.SecureWriteCloser) *MonitoredWriter {
	m := &MonitoredWriter{
		SourceField: net.NewSourceField(nil),
	}
	m.OutputField = net.NewOutputField(m, w)

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
