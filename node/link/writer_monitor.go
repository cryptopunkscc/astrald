package link

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/net"
)

var _ net.SecureWriteCloser = &WriterMonitor{}

type WriterMonitor struct {
	Target     net.SecureWriteCloser
	AfterWrite func(int, error)
	AfterClose func(error)
}

func (w *WriterMonitor) Write(p []byte) (n int, err error) {
	n, err = w.Target.Write(p)
	if w.AfterWrite != nil {
		w.AfterWrite(n, err)
	}
	return
}

func (w *WriterMonitor) Close() (err error) {
	err = w.Target.Close()
	if w.AfterClose != nil {
		w.AfterClose(err)
	}
	return
}

func (w *WriterMonitor) RemoteIdentity() id.Identity {
	return w.Target.RemoteIdentity()
}
