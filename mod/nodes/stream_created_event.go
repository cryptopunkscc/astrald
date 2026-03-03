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

func (e StreamCreatedEvent) ObjectType() string { return "mod.nodes.stream_created_event" }

func (e StreamCreatedEvent) WriteTo(w io.Writer) (int64, error) {
	return astral.Objectify(&e).WriteTo(w)
}

func (e *StreamCreatedEvent) ReadFrom(r io.Reader) (int64, error) {
	return astral.Objectify(e).ReadFrom(r)
}

func (e StreamCreatedEvent) MarshalJSON() ([]byte, error) {
	type alias StreamCreatedEvent
	return json.Marshal(alias(e))
}

func (e *StreamCreatedEvent) UnmarshalJSON(b []byte) error {
	type alias StreamCreatedEvent
	var a alias
	if err := json.Unmarshal(b, &a); err != nil {
		return err
	}
	*e = StreamCreatedEvent(a)
	return nil
}

func init() {
	_ = astral.Add(&StreamCreatedEvent{})
}
