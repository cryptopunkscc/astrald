package router

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/net"
	"sync"
)

type LockableWriter struct {
	*net.OutputField
	*net.SourceField
	sync.Mutex
	removed bool
}

func NewLockableWriter(output net.SecureWriteCloser) *LockableWriter {
	w := &LockableWriter{
		SourceField: net.NewSourceField(nil),
	}
	w.OutputField = net.NewOutputField(w, output)
	return w
}

func (w *LockableWriter) Identity() id.Identity {
	w.Lock()
	defer w.Unlock()

	return w.Output().Identity()
}

func (w *LockableWriter) Write(p []byte) (n int, err error) {
	w.Lock()
	defer w.Unlock()

	if p != nil {
		n, err = w.Output().Write(p)
		if err != nil {
			return n, err
		}
	}

	if !w.removed {
		s, ok := w.Source().(net.OutputSetter)
		if !ok {
			panic("not good")
		}
		s.SetOutput(w.Output())
		if o, ok := w.Output().(net.SourceSetter); ok {
			o.SetSource(s)
		}

		w.removed = true
	}

	return
}

func (w *LockableWriter) Close() error {
	w.Lock()
	defer w.Unlock()

	return w.Output().Close()
}
