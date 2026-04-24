package views

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/fmt"
	"github.com/cryptopunkscc/astrald/mod/log/theme"
)

type NilView struct {
	astral.Nil
}

func (NilView) Render() string {
	t := theme.Nil
	p := t.Bri(theme.Least)
	return p.Render("(") + t.Render("nil") + p.Render(")")
}

func init() {
	fmt.SetView(func(*astral.Nil) fmt.View {
		return &NilView{}
	})
}
