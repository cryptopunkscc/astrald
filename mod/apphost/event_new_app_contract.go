package apphost

import (
	"io"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/auth"
)

type EventNewAppContract struct {
	Contract *auth.SignedContract
}

var _ astral.Object = &EventNewAppContract{}

func (e EventNewAppContract) ObjectType() string {
	return "mod.apphost.events.new_app_contract"
}

func (e EventNewAppContract) WriteTo(w io.Writer) (n int64, err error) {
	return astral.Objectify(&e).WriteTo(w)
}

func (e *EventNewAppContract) ReadFrom(r io.Reader) (n int64, err error) {
	return astral.Objectify(e).ReadFrom(r)
}

func init() {
	astral.Add(&EventNewAppContract{})
}
