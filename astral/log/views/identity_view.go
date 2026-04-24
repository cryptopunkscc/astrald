package views

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/fmt"
)

type IdentityView struct {
	*astral.Identity
}

func (v IdentityView) Render() string {
	return v.Identity.Fingerprint()
}

func init() {
	fmt.SetView(func(o *astral.Identity) fmt.View {
		return IdentityView{Identity: o}
	})
}
