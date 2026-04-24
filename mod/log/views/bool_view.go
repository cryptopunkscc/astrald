package views

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/fmt"
	"github.com/cryptopunkscc/astrald/mod/log/theme"
)

type BoolView struct {
	*astral.Bool
}

func (b BoolView) Render() string {
	if *b.Bool {
		return theme.True.Render(b.Bool.String())
	}

	return theme.False.Render(b.Bool.String())
}

func init() {
	fmt.SetView(func(o *astral.Bool) fmt.View {
		return &BoolView{Bool: o}
	})
}
