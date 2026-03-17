package gateway

import (
	"io"

	"github.com/cryptopunkscc/astrald/astral"
)

var _ astral.Object = &Ping{}

type Ping struct {
	Pong astral.Bool
}

func (Ping) ObjectType() string                     { return "mod.gateway.ping" }
func (p Ping) WriteTo(w io.Writer) (int64, error)   { return astral.Objectify(&p).WriteTo(w) }
func (p *Ping) ReadFrom(r io.Reader) (int64, error) { return astral.Objectify(p).ReadFrom(r) }

var _ astral.Object = &Handoff{}

type Handoff struct{}

func (Handoff) ObjectType() string                     { return "mod.gateway.handoff" }
func (s Handoff) WriteTo(w io.Writer) (int64, error)   { return astral.Objectify(&s).WriteTo(w) }
func (s *Handoff) ReadFrom(r io.Reader) (int64, error) { return astral.Objectify(s).ReadFrom(r) }

var _ astral.Object = &HandoffAck{}

type HandoffAck struct{}

func (HandoffAck) ObjectType() string                     { return "mod.gateway.handoff_ack" }
func (s HandoffAck) WriteTo(w io.Writer) (int64, error)   { return astral.Objectify(&s).WriteTo(w) }
func (s *HandoffAck) ReadFrom(r io.Reader) (int64, error) { return astral.Objectify(s).ReadFrom(r) }

func init() { _ = astral.Add(&Handoff{}) }
func init() { _ = astral.Add(&HandoffAck{}) }
func init() { _ = astral.Add(&Ping{}) }
