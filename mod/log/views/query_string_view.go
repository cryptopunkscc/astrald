package views

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/log"
	"github.com/cryptopunkscc/astrald/lib/query"
	"github.com/cryptopunkscc/astrald/mod/log/styles"
)

type QueryStringView struct {
	*astral.String16
}

func NewQueryStringView(str string) QueryStringView {
	return QueryStringView{astral.NewString16(str)}
}

func (view QueryStringView) Render() (out string) {
	op, params := query.Parse(view.String16.String())

	out = String(op, &styles.GreenText).Render()

	if len(params) > 0 {
		out += String("?", &styles.GrayText).Render()
	}

	var first = true
	for name, field := range params {
		if !first {
			out += String("&", &styles.DarkGrayText).Render()
		}
		out += log.Render(
			String(name, &styles.DarkYellowText),
			String("=", &styles.GrayText),
			String(field, &styles.YellowText),
		)
		first = false
	}

	return out
}
