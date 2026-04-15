package log

import "github.com/cryptopunkscc/astrald/astral"

type QueryView struct {
	*astral.Query
}

func (view QueryView) Render() string {
	return view.QueryString
}

func init() {
	DefaultViewer.Set(astral.Query{}.ObjectType(), func(o astral.Object) astral.Object {
		return QueryView{Query: o.(*astral.Query)}
	})
}
