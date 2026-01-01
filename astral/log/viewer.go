package log

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/sig"
)

var DefaultViewer = &Viewer{}

type Viewer struct {
	viewers sig.Map[string, func(astral.Object) astral.Object]
}

func (v *Viewer) Render(objects ...astral.Object) (s string) {
	line := v.View(objects...)
	return Render(line...)
}

func (v *Viewer) View(obj ...astral.Object) (views []astral.Object) {
	for _, o := range obj {
		if w, found := v.viewers.Get(o.ObjectType()); found {
			views = append(views, w(o))
		} else {
			views = append(views, o)
		}
	}
	return
}

func (v *Viewer) Set(typ string, fn func(astral.Object) astral.Object) {
	if fn == nil {
		v.viewers.Delete(typ)
	} else {
		v.viewers.Replace(typ, fn)
	}
}

func (v *Viewer) Types() []string {
	return v.viewers.Keys()
}
