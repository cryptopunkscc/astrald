package ops

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/log"
	"github.com/cryptopunkscc/astrald/mod/log/styles"
	"github.com/cryptopunkscc/astrald/mod/log/views"
)

type OpSpecView struct {
	*OpSpec
}

func (op OpSpecView) Render() (out string) {
	view := log.DefaultViewer

	// name(
	out += view.Render(views.String(op.Name, &styles.BrightGreenText))
	out += view.Render(views.String("(", &styles.WhiteText))

	var first = true
	for _, spec := range op.Parameters {
		if !first {
			out += ", "
		}
		req := ""
		if spec.Required {
			req = "*"
		}
		out += view.Render(
			views.String(spec.Name+" ", &styles.GrayText),
			views.String(req, &styles.RedText),
			views.String(spec.Type, &styles.YellowText),
		)
		first = false
	}

	// )
	out += view.Render(views.String(")", &styles.WhiteText))

	return
}

func init() {
	log.DefaultViewer.Set(OpSpec{}.ObjectType(), func(object astral.Object) astral.Object {
		return &OpSpecView{OpSpec: object.(*OpSpec)}
	})
}
