package views

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/fmt"
)

type QueryView struct {
	*astral.Query
}

func (view QueryView) Render() string {
	return fmt.Sprintf(
		"[%v] %v -> %v:%v",
		&view.Nonce,
		view.Caller,
		view.Target,
		view.QueryString,
	)
}

func init() {
	fmt.SetView(func(o *astral.Query) fmt.View {
		return QueryView{Query: o}
	})
}
