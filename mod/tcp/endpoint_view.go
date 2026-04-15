package tcp

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/log"
	"github.com/cryptopunkscc/astrald/mod/log/styles"
)

type EndpointView struct {
	*Endpoint
}

func (v *EndpointView) Render() string {
	return styles.BlueText.Render("tcp:") +
		styles.BrightBlueText.Render(v.String())
}

func init() {
	log.DefaultViewer.Set(Endpoint{}.ObjectType(), func(object astral.Object) astral.Object {
		return &EndpointView{object.(*Endpoint)}
	})
}
