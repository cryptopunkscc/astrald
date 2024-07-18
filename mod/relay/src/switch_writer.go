package relay

import (
	"github.com/cryptopunkscc/astrald/id"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/streams"
	"sync/atomic"
)

type SwitchWriter struct {
	*astral.OutputField
	*astral.SourceField
	NextWriter  *LockableWriter
	AfterSwitch func()
	switched    atomic.Bool
}

func NewSwitchWriter(output astral.SecureWriteCloser) *SwitchWriter {
	w := &SwitchWriter{
		SourceField: astral.NewSourceField(nil),
		NextWriter:  NewLockableWriter(nil),
	}
	w.OutputField = astral.NewOutputField(w, output)
	w.NextWriter.OutputField = astral.NewOutputField(w, output)
	w.NextWriter.Lock()

	return w
}

func (w *SwitchWriter) Identity() id.Identity {
	return w.Output().Identity()
}

func (w *SwitchWriter) Write(p []byte) (n int, err error) {
	return w.Output().Write(p)
}

func (w *SwitchWriter) Close() error {
	if w.switched.CompareAndSwap(false, true) {
		w.SetOutput(astral.NewSecurePipeWriter(streams.NilWriteCloser{}, id.Identity{}))

		if s, ok := w.NextWriter.Output().(astral.SourceSetter); ok {
			s.SetSource(w.NextWriter)
		}

		if w.AfterSwitch != nil {
			w.AfterSwitch()
		}
		w.NextWriter.Unlock()
		w.NextWriter.Write(nil)
	}
	return nil
}
