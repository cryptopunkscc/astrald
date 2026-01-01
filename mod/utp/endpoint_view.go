package utp

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/log"
	modlog "github.com/cryptopunkscc/astrald/mod/log"
)

type EndpointView struct {
	*Endpoint
}

func (v *EndpointView) Render() string {
	return modlog.BlueText.Render("utp:") +
		modlog.BrightBlueText.Render(v.String())
}

func init() {
	log.DefaultViewer.Set(Endpoint{}.ObjectType(), func(object astral.Object) astral.Object {
		return &EndpointView{object.(*Endpoint)}
	})
}
