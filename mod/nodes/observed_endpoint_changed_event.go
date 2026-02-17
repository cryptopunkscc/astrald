package nodes

import (
	"encoding/json"
	"io"

	"github.com/cryptopunkscc/astrald/astral"
)

type ObservedEndpointChangedEvent struct{}

func (m ObservedEndpointChangedEvent) ObjectType() string {
	return "mod.nodes.observed_endpoint_changed_event"
}

func (m ObservedEndpointChangedEvent) WriteTo(w io.Writer) (int64, error) {
	return astral.Struct(m).WriteTo(w)
}

func (m *ObservedEndpointChangedEvent) ReadFrom(r io.Reader) (int64, error) {
	return astral.Struct(m).ReadFrom(r)
}

func (m ObservedEndpointChangedEvent) MarshalJSON() ([]byte, error) {
	type alias ObservedEndpointChangedEvent
	return json.Marshal(alias(m))
}

func (m *ObservedEndpointChangedEvent) UnmarshalJSON(b []byte) error {
	type alias ObservedEndpointChangedEvent
	var a alias
	if err := json.Unmarshal(b, &a); err != nil {
		return err
	}
	*m = ObservedEndpointChangedEvent(a)
	return nil
}

func init() {
	_ = astral.Add(&ObservedEndpointChangedEvent{})
}
