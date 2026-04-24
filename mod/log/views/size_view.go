package views

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/fmt"
	"github.com/cryptopunkscc/astrald/mod/log/theme"
)

type SizeView struct {
	*astral.Size
}

func (view SizeView) Render() string {
	b := view.Size.HumanReadableBinary()
	c := theme.Size

	var s, u string

	if b[len(b)-2] == 'i' {
		s, u = b[:len(b)-3], b[len(b)-3:]
	} else {
		s, u = b[:len(b)-1], b[len(b)-1:]
	}

	return c.Render(s) + c.Bri(theme.Least).Render(u)
}

func init() {
	fmt.SetView(func(o *astral.Size) fmt.View {
		return &SizeView{Size: o}
	})
}
