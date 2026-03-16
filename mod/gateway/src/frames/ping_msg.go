package frames

import (
	"io"

	"github.com/cryptopunkscc/astrald/astral"
)

type PingMsg struct{}

func (PingMsg) ObjectType() string                   { return "mod.gateway.ping" }
func (PingMsg) WriteTo(w io.Writer) (int64, error)   { return 0, nil }
func (*PingMsg) ReadFrom(r io.Reader) (int64, error) { return 0, nil }

func init() { _ = astral.Add(&PingMsg{}) }
