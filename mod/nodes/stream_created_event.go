package nodes

import (
	"encoding/json"
	"io"

	"github.com/cryptopunkscc/astrald/astral"
)

type StreamCreatedEvent struct {
	RemoteIdentity *astral.Identity
	StreamId       astral.Nonce
	StreamCount    int
}

func (m StreamCreatedEvent) ObjectType() string { return "mod.nodes.stream_created_event" }

func (m StreamCreatedEvent) WriteTo(w io.Writer) (int64, error) { return astral.Struct(m).WriteTo(w) }

func (m *StreamCreatedEvent) ReadFrom(r io.Reader) (int64, error) {
	return astral.Struct(m).ReadFrom(r)
}

func (m StreamCreatedEvent) MarshalJSON() ([]byte, error) {
	type alias StreamCreatedEvent
	return json.Marshal(alias(m))
}

func (m *StreamCreatedEvent) UnmarshalJSON(b []byte) error {
	type alias StreamCreatedEvent
	var a alias
	if err := json.Unmarshal(b, &a); err != nil {
		return err
	}
	*m = StreamCreatedEvent(a)
	return nil
}

func init() {
	_ = astral.Add(&StreamCreatedEvent{})
}
