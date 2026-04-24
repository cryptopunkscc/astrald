package views

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/fmt"
	"github.com/cryptopunkscc/astrald/mod/log/theme"
)

type EOSView struct {
	astral.EOS
}

func (EOSView) Render() string {
	t := theme.EOS
	p := t.Bri(theme.Least)
	return p.Render("(") + t.Render("end of stream") + p.Render(")")
}

func init() {
	fmt.SetView(func(*astral.EOS) fmt.View {
		return &EOSView{}
	})
}
