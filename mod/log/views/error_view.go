package views

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/fmt"
	"github.com/cryptopunkscc/astrald/mod/log/theme"
)

type ErrorView struct {
	astral.Error
}

func (v ErrorView) Render() string {
	return theme.Error.Render(v.Error.Error())
}

func init() {
	fmt.SetView(func(o *astral.ErrorMessage) fmt.View {
		return ErrorView{Error: o}
	})
}
