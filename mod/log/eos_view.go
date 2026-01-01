package log

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/log"
)

type EOSView struct {
	astral.EOS
}

func (EOSView) Render() string {
	return DarkGrayText.Render("(end of stream)")
}

func init() {
	log.DefaultViewer.Set(astral.EOS{}.ObjectType(), func(object astral.Object) astral.Object {
		return &EOSView{}
	})
}
