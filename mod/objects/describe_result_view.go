package objects

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/log"
)

type DescribeResultView struct {
	*DescribeResult
}

func (view DescribeResult) Render() string {
	return log.DefaultViewer.Render(log.Format("%v %v",
		"➤",
		view.Descriptor,
	)...)
}

func init() {
	log.DefaultViewer.Set(DescribeResult{}.ObjectType(), func(object astral.Object) astral.Object {
		return &DescribeResultView{object.(*DescribeResult)}
	})
}
