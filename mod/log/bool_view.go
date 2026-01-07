package log

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/log"
)

type BoolView struct {
	*astral.Bool
}

func (b BoolView) Render() string {
	if *b.Bool {
		return GreenText.Render(b.Bool.String())
	}

	return RedText.Render(b.Bool.String())
}

func init() {
	log.DefaultViewer.Set(astral.Bool(false).ObjectType(), func(object astral.Object) astral.Object {
		return &BoolView{Bool: object.(*astral.Bool)}
	})
}
