package frames

import (
	"io"

	"github.com/cryptopunkscc/astrald/astral"
)

type SignalReadyMsg struct{}

func (SignalReadyMsg) ObjectType() string                   { return "mod.gateway.signal_ready" }
func (SignalReadyMsg) WriteTo(w io.Writer) (int64, error)   { return 0, nil }
func (*SignalReadyMsg) ReadFrom(r io.Reader) (int64, error) { return 0, nil }

func init() { _ = astral.Add(&SignalReadyMsg{}) }
