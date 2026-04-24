package views

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/fmt"
)

type QueryView struct {
	*astral.Query
}

func (view QueryView) Render() (out string) {
	out = fmt.Sprintf(
		"[%v] %v -> %v:%v",
		&view.Nonce,
		view.Caller,
		view.Target,
		NewQueryStringView(view.QueryString),
	)

	return out
}

func UseQueryView() {
	fmt.SetView(func(o *astral.Query) fmt.View {
		return QueryView{o}
	})
}
