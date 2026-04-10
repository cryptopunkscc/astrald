package log

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/log"
	"github.com/cryptopunkscc/astrald/lib/query"
)

type QueryStringView struct {
	*astral.String16
}

func NewQueryStringView(str string) QueryStringView {
	return QueryStringView{astral.NewString16(str)}
}

func (view QueryStringView) Render() (out string) {
	op, params := query.Parse(view.String16.String())

	out = String(op, &GreenText).Render()

	if len(params) > 0 {
		out += String("?", &GrayText).Render()
	}

	var first = true
	for name, field := range params {
		if !first {
			out += String("&", &DarkGrayText).Render()
		}
		out += log.Render(
			String(name, &DarkYellowText),
			String("=", &GrayText),
			String(field, &YellowText),
		)
		first = false
	}

	return out
}
