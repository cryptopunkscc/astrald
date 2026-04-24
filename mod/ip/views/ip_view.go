package ip

import (
	"github.com/cryptopunkscc/astrald/astral/fmt"
	"github.com/cryptopunkscc/astrald/mod/ip"
	"github.com/cryptopunkscc/astrald/mod/log/theme"
)

type IPView struct {
	*ip.IP
}

func (v IPView) Render() string {
	return theme.Tertiary.Render(v.IP.String())
}

func init() {
	fmt.SetView(func(o *ip.IP) fmt.View {
		return &IPView{IP: o}
	})
}
