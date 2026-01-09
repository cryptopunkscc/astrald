package nodes

import (
	"encoding/json"
	"io"

	"github.com/cryptopunkscc/astrald/astral"
)

type StreamClosedEvent struct {
	RemoteIdentity *astral.Identity // Identity of the other party
	Forced         astral.Bool
	StreamCount    astral.Int8
}

func (StreamClosedEvent) ObjectType() string {
	return "mod.nodes.stream_closed_event"
}

func (e StreamClosedEvent) WriteTo(w io.Writer) (n int64, err error) {
	return astral.Struct(e).WriteTo(w)
}

func (e *StreamClosedEvent) ReadFrom(r io.Reader) (n int64, err error) {
	return astral.Struct(e).ReadFrom(r)
}

func (m StreamClosedEvent) MarshalJSON() ([]byte, error) {
	type alias StreamClosedEvent
	return json.Marshal(alias(m))
}

func (m *StreamClosedEvent) UnmarshalJSON(b []byte) error {
	type alias StreamClosedEvent
	var a alias
	if err := json.Unmarshal(b, &a); err != nil {
		return err
	}
	*m = StreamClosedEvent(a)
	return nil
}

func init() {
	_ = astral.DefaultBlueprints.Add(&StreamClosedEvent{})
}
