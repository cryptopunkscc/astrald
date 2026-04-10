package ops

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/log"
	modlog "github.com/cryptopunkscc/astrald/mod/log"
)

type OpSpecView struct {
	*OpSpec
}

func (op OpSpecView) Render() (out string) {
	view := log.DefaultViewer

	// name(
	out += view.Render(modlog.String(op.Name, &modlog.BrightGreenText))
	out += view.Render(modlog.String("(", &modlog.WhiteText))

	var first = true
	for arg, spec := range op.Parameters {
		if !first {
			out += ", "
		}
		req := ""
		if spec.Required {
			req = "*"
		}
		out += view.Render(
			modlog.String(arg+" ", &modlog.GrayText),
			modlog.String(req, &modlog.RedText),
			modlog.String(spec.Type, &modlog.YellowText),
		)
		first = false
	}

	// )
	out += view.Render(modlog.String(")", &modlog.WhiteText))

	return
}

func init() {
	log.DefaultViewer.Set(OpSpec{}.ObjectType(), func(object astral.Object) astral.Object {
		return &OpSpecView{OpSpec: object.(*OpSpec)}
	})
}
