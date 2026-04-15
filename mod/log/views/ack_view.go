package views

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/log"
	"github.com/cryptopunkscc/astrald/mod/log/styles"
)

type AckView struct {
	astral.Ack
}

func (AckView) Render() string {
	return styles.DarkGreenText.Render("(ack)")
}

func init() {
	log.DefaultViewer.Set(astral.Ack{}.ObjectType(), func(object astral.Object) astral.Object {
		return &AckView{}
	})
}
