package log

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/log"
)

type ObjectIDView struct {
	*astral.ObjectID
}

func (v ObjectIDView) Render() string {
	return BlueText.Render(v.ObjectID.String())
}

func init() {
	log.DefaultViewer.Set(astral.ObjectID{}.ObjectType(), func(o astral.Object) astral.Object {
		return ObjectIDView{o.(*astral.ObjectID)}
	})
}
