package views

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/fmt"
	"github.com/cryptopunkscc/astrald/mod/dir"
	"github.com/cryptopunkscc/astrald/mod/log/theme"
	"github.com/cryptopunkscc/astrald/sig"
)

var IdentityResolver sig.Value[dir.Resolver]

type IdentityView struct {
	Highlight bool
	*astral.Identity
}

func (v IdentityView) Render() string {
	s := theme.Identity
	if v.Highlight {
		s = s.Bri(theme.More)
	}

	if r := IdentityResolver.Get(); r != nil {
		return s.Render(r.DisplayName(v.Identity))
	}

	return s.Render(v.Identity.Fingerprint())
}

func UseIdentityView() {
	fmt.SetView(func(identity *astral.Identity) fmt.View {
		return IdentityView{Identity: identity}
	})
}
