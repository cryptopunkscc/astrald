package frames

import (
	"io"

	"github.com/cryptopunkscc/astrald/astral"
)

type SignalGoMsg struct{}

func (SignalGoMsg) ObjectType() string                   { return "mod.gateway.signal_go" }
func (SignalGoMsg) WriteTo(w io.Writer) (int64, error)   { return 0, nil }
func (*SignalGoMsg) ReadFrom(r io.Reader) (int64, error) { return 0, nil }

func init() { _ = astral.Add(&SignalGoMsg{}) }
