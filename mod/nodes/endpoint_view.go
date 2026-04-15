package nodes

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/log"
	"github.com/cryptopunkscc/astrald/mod/log/styles"
)

type EndpointView struct {
	*EndpointWithTTL
}

func (v *EndpointView) Render() string {
	return styles.BlueText.Render(v.Network()+":") +
		styles.BrightBlueText.Render(v.Address())
}

func init() {
	log.DefaultViewer.Set(EndpointWithTTL{}.ObjectType(), func(object astral.Object) astral.Object {
		return &EndpointView{object.(*EndpointWithTTL)}
	})
}
