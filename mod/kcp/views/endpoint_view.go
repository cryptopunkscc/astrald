package kcp

import (
	"github.com/cryptopunkscc/astrald/astral/fmt"
	"github.com/cryptopunkscc/astrald/mod/kcp"
	"github.com/cryptopunkscc/astrald/mod/log/theme"
)

type EndpointView struct {
	*kcp.Endpoint
}

func (v *EndpointView) Render() string {
	n := theme.Tertiary
	b := n.Bri(theme.More)

	return n.Render("kcp:") + b.Render(v.Address())
}

func init() {
	fmt.SetView(func(o *kcp.Endpoint) fmt.View {
		return &EndpointView{Endpoint: o}
	})
}
