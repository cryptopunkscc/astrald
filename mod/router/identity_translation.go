package router

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/net"
)

type IdentityTranslation struct {
	*net.OutputField
	*net.SourceField
	net.SecureWriteCloser
	identity id.Identity
}

func NewIdentityTranslation(w net.SecureWriteCloser, identity id.Identity) *IdentityTranslation {
	var t = &IdentityTranslation{
		SourceField:       net.NewSourceField(nil),
		SecureWriteCloser: w,
		identity:          identity,
	}
	t.OutputField = net.NewOutputField(t, w)
	return t
}

func (r *IdentityTranslation) Identity() id.Identity {
	return r.identity
}
