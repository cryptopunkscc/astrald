package nodes

import (
	"io"

	"github.com/cryptopunkscc/astrald/astral"
)

type StreamPressureState struct {
	Level astral.Float64  // current bucket fill after decay (bytes)
	RTT   astral.Duration // smoothed round-trip time (EMA)
	Score astral.Float64  // combined weighted score (0..~3+)
	High  astral.Bool     // true while score is above the Enter threshold
}

var _ astral.Object = &StreamPressureState{}

func (StreamPressureState) ObjectType() string { return "mod.nodes.stream_pressure_state" }

func (s StreamPressureState) WriteTo(w io.Writer) (int64, error) {
	return astral.Objectify(&s).WriteTo(w)
}

func (s *StreamPressureState) ReadFrom(r io.Reader) (int64, error) {
	return astral.Objectify(s).ReadFrom(r)
}

func init() {
	_ = astral.Add(&StreamPressureState{})
}
