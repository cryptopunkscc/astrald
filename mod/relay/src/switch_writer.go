package relay

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/net"
	"github.com/cryptopunkscc/astrald/streams"
	"sync/atomic"
)

type SwitchWriter struct {
	*net.OutputField
	*net.SourceField
	NextWriter  *LockableWriter
	AfterSwitch func()
	switched    atomic.Bool
}

func NewSwitchWriter(output net.SecureWriteCloser) *SwitchWriter {
	w := &SwitchWriter{
		SourceField: net.NewSourceField(nil),
		NextWriter:  NewLockableWriter(nil),
	}
	w.OutputField = net.NewOutputField(w, output)
	w.NextWriter.OutputField = net.NewOutputField(w, output)
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
		w.SetOutput(net.NewSecurePipeWriter(streams.NilWriteCloser{}, id.Identity{}))

		if s, ok := w.NextWriter.Output().(net.SourceSetter); ok {
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
