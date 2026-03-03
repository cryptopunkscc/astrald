package nodes

import (
	"encoding/json"
	"io"

	"github.com/cryptopunkscc/astrald/astral"
)

type StreamPressureEvent struct {
	RemoteIdentity *astral.Identity
	StreamID       astral.Nonce
}

func (StreamPressureEvent) ObjectType() string { return "mod.nodes.stream_pressure_event" }

func (e StreamPressureEvent) WriteTo(w io.Writer) (n int64, err error) {
	return astral.Objectify(&e).WriteTo(w)
}

func (e *StreamPressureEvent) ReadFrom(r io.Reader) (n int64, err error) {
	return astral.Objectify(e).ReadFrom(r)
}

func (e StreamPressureEvent) MarshalJSON() ([]byte, error) {
	type alias StreamPressureEvent
	return json.Marshal(alias(e))
}

func (e *StreamPressureEvent) UnmarshalJSON(b []byte) error {
	type alias StreamPressureEvent
	var a alias
	if err := json.Unmarshal(b, &a); err != nil {
		return err
	}
	*e = StreamPressureEvent(a)
	return nil
}

func init() {
	_ = astral.Add(&StreamPressureEvent{})
}
