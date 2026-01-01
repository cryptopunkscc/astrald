package nat

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/log"
	modlog "github.com/cryptopunkscc/astrald/mod/log"
)

type UDPEndpointView struct {
	*UDPEndpoint
}

func (v *UDPEndpointView) Render() string {
	return modlog.BlueText.Render("udp:") +
		modlog.BrightBlueText.Render(v.String())
}

func init() {
	log.DefaultViewer.Set(UDPEndpoint{}.ObjectType(), func(object astral.Object) astral.Object {
		return &UDPEndpointView{object.(*UDPEndpoint)}
	})
}
