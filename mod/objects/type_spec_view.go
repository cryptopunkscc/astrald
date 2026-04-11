package objects

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/log"
	"github.com/cryptopunkscc/astrald/mod/log/styles"
	"github.com/cryptopunkscc/astrald/mod/log/views"
)

type TypeSpecView struct {
	*TypeSpec
}

func (t TypeSpecView) Render() (out string) {
	view := log.DefaultViewer

	// name{
	out += view.Render(views.String(t.Name, &styles.BrightGreenText))
	out += view.Render(views.String("{", &styles.WhiteText))

	var first = true
	for _, spec := range t.Fields {
		if !first {
			out += ", "
		}

		out += view.Render(
			views.String(spec.Name+" ", &styles.GrayText),
			views.String(spec.Type, &styles.YellowText),
		)
		first = false
	}

	// }
	out += view.Render(views.String("}", &styles.WhiteText))

	return
}

func init() {
	log.DefaultViewer.Set(TypeSpec{}.ObjectType(), func(object astral.Object) astral.Object {
		return &TypeSpecView{TypeSpec: object.(*TypeSpec)}
	})
}
