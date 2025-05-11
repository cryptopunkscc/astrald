package apphost

import (
	"github.com/cryptopunkscc/astrald/astral"
	"io"
)

type EventNewAppContract struct {
	Contract *AppContract
}

var _ astral.Object = &EventNewAppContract{}

func (e EventNewAppContract) ObjectType() string {
	return "mod.apphost.events.new_app_contract"
}

func (e EventNewAppContract) WriteTo(w io.Writer) (n int64, err error) {
	return astral.Struct(e).WriteTo(w)
}

func (e *EventNewAppContract) ReadFrom(r io.Reader) (n int64, err error) {
	return astral.Struct(e).ReadFrom(r)
}

func init() {
	astral.DefaultBlueprints.Add(&EventNewAppContract{})
}
