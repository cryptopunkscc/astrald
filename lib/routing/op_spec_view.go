package routing

import (
	"github.com/cryptopunkscc/astrald/astral/fmt"
	"github.com/cryptopunkscc/astrald/mod/log/styles"
	"github.com/cryptopunkscc/astrald/mod/log/theme"
)

type OpSpecView struct {
	*OpSpec
}

func (op OpSpecView) Render() (out string) {
	arg := theme.Normal
	sep := theme.Normal.Bri(theme.More)

	// name(
	out += theme.Op.Render(op.Name)
	out += sep.Render("(")

	var first = true
	for _, spec := range op.Parameters {
		if !first {
			out += ", "
		}
		req := ""
		if spec.Required {
			req = "*"
		}
		out += arg.Render(spec.Name) + " " +
			styles.Red.Render(req) +
			theme.Type.Render(spec.Type)
		first = false
	}

	// )
	out += sep.Render(")")

	return
}

func init() {
	fmt.SetView(func(o *OpSpec) fmt.View {
		return &OpSpecView{OpSpec: o}
	})
}
