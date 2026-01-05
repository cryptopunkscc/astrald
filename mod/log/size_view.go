package log

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/log"
)

type SizeView struct {
	*astral.Size
}

func (view SizeView) Render() string {
	return DarkYellowText.Render(view.Size.HumanReadableBinary())
}

func init() {
	log.DefaultViewer.Set(astral.Size(0).ObjectType(), func(object astral.Object) astral.Object {
		return &SizeView{object.(*astral.Size)}
	})
}
