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

	out = String(op, &styles.BrightYellowText).Render()

	if len(params) > 0 {
		out += String("?", &styles.DarkGrayText).Render()
	}

	var first = true
	for name, field := range params {
		if !first {
			out += String("&", &styles.DarkGrayText).Render()
		}
		out += log.Render(
			String(name, &styles.WhiteText),
			String("=", &styles.DarkGrayText),
			String(field, &styles.GrayText),
		)
		first = false
	}

	return out
}
