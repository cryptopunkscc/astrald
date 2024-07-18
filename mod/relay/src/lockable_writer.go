package relay

import (
	"github.com/cryptopunkscc/astrald/id"
	"github.com/cryptopunkscc/astrald/astral"
	"sync"
)

type LockableWriter struct {
	*astral.OutputField
	*astral.SourceField
	sync.Mutex
	removed bool
}

func NewLockableWriter(output astral.SecureWriteCloser) *LockableWriter {
	w := &LockableWriter{
		SourceField: astral.NewSourceField(nil),
	}
	w.OutputField = astral.NewOutputField(w, output)
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
		s, ok := w.Source().(astral.OutputSetter)
		if !ok {
			panic("not good")
		}
		s.SetOutput(w.Output())
		if o, ok := w.Output().(astral.SourceSetter); ok {
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
