package log

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/log"
)

type AckView struct {
	astral.Ack
}

func (AckView) Render() string {
	return DarkGreenText.Render("(ack)")
}

func init() {
	log.DefaultViewer.Set(astral.Ack{}.ObjectType(), func(object astral.Object) astral.Object {
		return &AckView{}
	})
}
