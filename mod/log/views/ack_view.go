package views

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/fmt"
	"github.com/cryptopunkscc/astrald/mod/log/theme"
)

type AckView struct {
	astral.Ack
}

func (AckView) Render() string {
	t := theme.Ack
	a := t.Bri(theme.Least)

	return a.Render("(") + t.Render("ack") + a.Render(")")
}

func init() {
	fmt.SetView(func(*astral.Ack) fmt.View {
		return &AckView{}
	})
}
