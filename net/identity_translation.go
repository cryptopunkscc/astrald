package net

import (
	"github.com/cryptopunkscc/astrald/id"
)

var _ SecureWriteCloser = &IdentityTranslation{}

type IdentityTranslation struct {
	*OutputField
	*SourceField
	identity id.Identity
}

func NewIdentityTranslation(w SecureWriteCloser, identity id.Identity) *IdentityTranslation {
	var t = &IdentityTranslation{
		SourceField: NewSourceField(nil),
		identity:    identity,
	}
	t.OutputField = NewOutputField(t, w)
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
