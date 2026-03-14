package gateway

import (
	"io"

	"github.com/cryptopunkscc/astrald/astral"
)

// PingFrame is exchanged between the gateway and a binder's idle socket conn.
// Ping=true is a ping (gateway→binder); Ping=false is a pong (binder→gateway).
// Stop=true signals the binder to stop the ping loop and proceed to link establishment.
type PingFrame struct {
	Ping astral.Bool
	Stop astral.Bool
}

func (PingFrame) ObjectType() string { return "mod.gateway.ping_frame" }

func (p PingFrame) WriteTo(w io.Writer) (int64, error)   { return astral.Objectify(&p).WriteTo(w) }
func (p *PingFrame) ReadFrom(r io.Reader) (int64, error) { return astral.Objectify(p).ReadFrom(r) }

func init() {
	astral.Add(&PingFrame{})
}
