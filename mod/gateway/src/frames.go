package gateway

import (
	"io"

	"github.com/cryptopunkscc/astrald/astral"
)

type Ping struct {
	Pong astral.Bool
}

func (Ping) ObjectType() string                     { return "mod.gateway.ping" }
func (p Ping) WriteTo(w io.Writer) (int64, error)   { return astral.Objectify(&p).WriteTo(w) }
func (p *Ping) ReadFrom(r io.Reader) (int64, error) { return astral.Objectify(p).ReadFrom(r) }

func init() { _ = astral.Add(&Ping{}) }

type Handoff struct {
	Confirm astral.Bool
}

func (Handoff) ObjectType() string                     { return "mod.gateway.signal" }
func (s Handoff) WriteTo(w io.Writer) (int64, error)   { return astral.Objectify(&s).WriteTo(w) }
func (s *Handoff) ReadFrom(r io.Reader) (int64, error) { return astral.Objectify(s).ReadFrom(r) }

func init() { _ = astral.Add(&Handoff{}) }
