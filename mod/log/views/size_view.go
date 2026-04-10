package views

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/log"
	"github.com/cryptopunkscc/astrald/mod/log/styles"
)

type SizeView struct {
	*astral.Size
}

func (view SizeView) Render() string {
	return styles.DarkYellowText.Render(view.Size.HumanReadableBinary())
}

func init() {
	log.DefaultViewer.Set(astral.Size(0).ObjectType(), func(object astral.Object) astral.Object {
		return &SizeView{object.(*astral.Size)}
	})
}
