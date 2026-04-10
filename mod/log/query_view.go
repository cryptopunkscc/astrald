package log

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/log"
)

type QueryView struct {
	*astral.Query
}

func (view QueryView) Render() (out string) {
	out = log.DefaultViewer.Render(log.Format(
		"[%v] %v -> %v:%v",
		view.Nonce,
		view.Caller,
		view.Target,
		NewQueryStringView(view.QueryString),
	)...)

	return out
}

func UseQueryView() {
	log.DefaultViewer.Set(astral.Query{}.ObjectType(), func(o astral.Object) astral.Object {
		return QueryView{Query: o.(*astral.Query)}
	})
}
