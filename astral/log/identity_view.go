package log

import "github.com/cryptopunkscc/astrald/astral"

type IdentityView struct {
	*astral.Identity
}

func (v IdentityView) Render() string {
	return v.Identity.Fingerprint()
}

func init() {
	DefaultViewer.Set(astral.Identity{}.ObjectType(), func(o astral.Object) astral.Object {
		return IdentityView{Identity: o.(*astral.Identity)}
	})
}
