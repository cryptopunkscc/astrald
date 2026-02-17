package nodes

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/log"
	modlog "github.com/cryptopunkscc/astrald/mod/log"
)

type EndpointView struct {
	*ResolvedEndpoint
}

func (v *EndpointView) Render() string {
	return modlog.BlueText.Render(v.Network()+":") +
		modlog.BrightBlueText.Render(v.Address())
}

func init() {
	log.DefaultViewer.Set(ResolvedEndpoint{}.ObjectType(), func(object astral.Object) astral.Object {
		return &EndpointView{object.(*ResolvedEndpoint)}
	})
}
