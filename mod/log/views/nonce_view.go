package views

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/fmt"
	"github.com/cryptopunkscc/astrald/mod/log/theme"
)

type NonceView struct {
	*astral.Nonce
}

func (v NonceView) Render() string {
	return theme.Nonce.Render(v.Nonce.String())
}

func init() {
	fmt.SetView(func(o *astral.Nonce) fmt.View {
		return &NonceView{Nonce: o}
	})
}
