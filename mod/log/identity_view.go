package log

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/log"
	"github.com/cryptopunkscc/astrald/mod/dir"
	"github.com/cryptopunkscc/astrald/sig"
)

var IdentityResolver sig.Value[dir.Resolver]

type IdentityView struct {
	Highlight bool
	*astral.Identity
}

func (v IdentityView) Render() string {
	var style = GreenText
	if v.Highlight {
		style = BrightGreenText
	}

	if r := IdentityResolver.Get(); r != nil {
		return style.Render(r.DisplayName(v.Identity))
	}

	return style.Render(v.Identity.Fingerprint())
}

func UseIdentityView() {
	log.DefaultViewer.Set(
		astral.Identity{}.ObjectType(),
		func(object astral.Object) astral.Object {
			return IdentityView{Identity: object.(*astral.Identity)}
		},
	)
}
