package router

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/net"
)

var _ net.SecureWriteCloser = &IdentityTranslation{}

type IdentityTranslation struct {
	*net.OutputField
	*net.SourceField
	identity id.Identity
}

func NewIdentityTranslation(w net.SecureWriteCloser, identity id.Identity) *IdentityTranslation {
	var t = &IdentityTranslation{
		SourceField: net.NewSourceField(nil),
		identity:    identity,
	}
	t.OutputField = net.NewOutputField(t, w)
	return t
}

func (r *IdentityTranslation) Write(p []byte) (n int, err error) {
	return r.Output().Write(p)
}

func (r *IdentityTranslation) Close() error {
	return r.Output().Close()
}

func (r *IdentityTranslation) Identity() id.Identity {
	return r.identity
}
