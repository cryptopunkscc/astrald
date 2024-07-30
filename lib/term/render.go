package term

import "github.com/cryptopunkscc/astrald/sig"

type Renderer interface {
	SetColor(Color)
	Text(string)
}

var _ Renderer = &RendererSet{}

type RendererSet struct {
	sig.Set[Renderer]
}

func (set *RendererSet) SetColor(color Color) {
	for _, r := range set.Set.Clone() {
		r.SetColor(color)
	}
}

func (set *RendererSet) Text(s string) {
	for _, r := range set.Set.Clone() {
		r.Text(s)
	}
}

var _ Renderer = &NilRenderer{}

type NilRenderer struct{}

func (n NilRenderer) SetColor(color Color) {
}

func (n NilRenderer) Text(s string) {
}
