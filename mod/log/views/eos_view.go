package views

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/log"
	"github.com/cryptopunkscc/astrald/mod/log/styles"
)

type EOSView struct {
	astral.EOS
}

func (EOSView) Render() string {
	return styles.DarkGrayText.Render("(end of stream)")
}

func init() {
	log.DefaultViewer.Set(astral.EOS{}.ObjectType(), func(object astral.Object) astral.Object {
		return &EOSView{}
	})
}
