package log

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/log"
)

type NilView struct {
	astral.Nil
}

func (NilView) Render() string {
	return RedText.Render("nil")
}

func init() {
	log.DefaultViewer.Set(astral.Nil{}.ObjectType(), func(object astral.Object) astral.Object {
		return &NilView{}
	})
}
