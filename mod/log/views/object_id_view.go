package views

import (
	"strings"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/fmt"
	"github.com/cryptopunkscc/astrald/mod/log/theme"
)

type ObjectIDView struct {
	*astral.ObjectID
}

func (v ObjectIDView) Render() string {
	t := theme.ObjectID
	p := t.Bri(theme.Least)
	str := v.ObjectID.String()
	str = strings.TrimPrefix(str, "data1")
	return p.Render("data1") + t.Render(str)
}

func init() {
	fmt.SetView(func(o *astral.ObjectID) fmt.View {
		return ObjectIDView{ObjectID: o}
	})
}
