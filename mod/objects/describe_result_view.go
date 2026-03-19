package objects

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/log"
)

type DescriptorView struct {
	*Descriptor
}

func (view Descriptor) Render() string {
	return log.DefaultViewer.Render(log.Format("%v %v",
		"➤",
		view.Data,
	)...)
}

func init() {
	log.DefaultViewer.Set(Descriptor{}.ObjectType(), func(object astral.Object) astral.Object {
		return &DescriptorView{object.(*Descriptor)}
	})
}
