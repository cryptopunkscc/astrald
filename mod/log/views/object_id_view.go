package views

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/log"
	"github.com/cryptopunkscc/astrald/mod/log/styles"
)

type ObjectIDView struct {
	*astral.ObjectID
}

func (v ObjectIDView) Render() string {
	return styles.BlueText.Render(v.ObjectID.String())
}

func init() {
	log.DefaultViewer.Set(astral.ObjectID{}.ObjectType(), func(o astral.Object) astral.Object {
		return ObjectIDView{o.(*astral.ObjectID)}
	})
}
