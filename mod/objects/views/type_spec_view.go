package objects

import (
	"github.com/cryptopunkscc/astrald/astral/fmt"
	"github.com/cryptopunkscc/astrald/mod/log/styles"
	"github.com/cryptopunkscc/astrald/mod/log/theme"
	"github.com/cryptopunkscc/astrald/mod/objects"
)

type TypeSpecView struct {
	*objects.TypeSpec
}

func (t TypeSpecView) Render() (out string) {
	// name{
	out += theme.Type.Bri(theme.Less).Render(t.Name)
	out += styles.Highlight.Render("{")

	var first = true
	for _, spec := range t.Fields {
		if !first {
			out += ", "
		}

		out += fmt.Sprint(
			styles.String(spec.Name+" ", theme.Normal),
			styles.String(spec.Type, theme.Type),
		)
		first = false
	}

	// }
	out += styles.Highlight.Render("}")

	return
}

func init() {
	fmt.SetView(func(o *objects.TypeSpec) fmt.View {
		return &TypeSpecView{TypeSpec: o}
	})
}
