package views

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/log"
	"github.com/cryptopunkscc/astrald/mod/log/styles"
)

type NilView struct {
	astral.Nil
}

func (NilView) Render() string {
	return styles.RedText.Render("nil")
}

func init() {
	log.DefaultViewer.Set(astral.Nil{}.ObjectType(), func(object astral.Object) astral.Object {
		return &NilView{}
	})
}
