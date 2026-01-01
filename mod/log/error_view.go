package log

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/log"
)

type ErrorView struct {
	astral.Error
}

func (v ErrorView) Render() string {
	return RedText.Render(v.Error.Error())
}

func init() {
	log.DefaultViewer.Set(astral.ErrorMessage{}.ObjectType(), func(object astral.Object) astral.Object {
		return ErrorView{object.(*astral.ErrorMessage)}
	})
}
