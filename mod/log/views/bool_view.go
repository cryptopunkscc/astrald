package views

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/log"
	"github.com/cryptopunkscc/astrald/mod/log/styles"
)

type BoolView struct {
	*astral.Bool
}

func (b BoolView) Render() string {
	if *b.Bool {
		return styles.GreenText.Render(b.Bool.String())
	}

	return styles.RedText.Render(b.Bool.String())
}

func init() {
	log.DefaultViewer.Set(astral.Bool(false).ObjectType(), func(object astral.Object) astral.Object {
		return &BoolView{Bool: object.(*astral.Bool)}
	})
}
