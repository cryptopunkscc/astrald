package gateway

import (
	"github.com/cryptopunkscc/astrald/astral/fmt"
	"github.com/cryptopunkscc/astrald/mod/gateway"
	"github.com/cryptopunkscc/astrald/mod/log/theme"
)

type EndpointView struct {
	*gateway.Endpoint
}

func (v *EndpointView) Render() string {
	n := theme.Tertiary
	b := n.Bri(theme.More)

	return n.Render("gw:") +
		b.Render(v.GatewayID.String()) +
		n.Render(":") +
		b.Render(v.TargetID.String())
}

func init() {
	fmt.SetView(func(o *gateway.Endpoint) fmt.View {
		return &EndpointView{Endpoint: o}
	})
}
