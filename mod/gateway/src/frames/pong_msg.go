package frames

import (
	"io"

	"github.com/cryptopunkscc/astrald/astral"
)

type PongMsg struct{}

func (PongMsg) ObjectType() string                   { return "mod.gateway.pong" }
func (PongMsg) WriteTo(w io.Writer) (int64, error)   { return 0, nil }
func (*PongMsg) ReadFrom(r io.Reader) (int64, error) { return 0, nil }

func init() { _ = astral.Add(&PongMsg{}) }
