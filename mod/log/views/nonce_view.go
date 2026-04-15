package views

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/log"
	"github.com/cryptopunkscc/astrald/mod/log/styles"
)

type NonceView struct {
	*astral.Nonce
}

func (v NonceView) Render() string {
	return styles.YellowText.Render(v.Nonce.String())
}

func init() {
	log.DefaultViewer.Set(astral.Nonce(0).ObjectType(), func(o astral.Object) astral.Object {
		return &NonceView{o.(*astral.Nonce)}
	})
}
