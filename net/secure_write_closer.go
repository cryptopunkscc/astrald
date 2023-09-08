package net

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	"io"
)

type Secure interface {
	Identity() id.Identity
}

type SecureWriteCloser interface {
	Secure
	io.WriteCloser
}

type SecureReader interface {
	Secure
	io.Reader
}

var _ SecureWriteCloser = &secureWriteCloser{}

type secureWriteCloser struct {
	*SourceField
	io.WriteCloser
	identity id.Identity
}

func NewSecureWriteCloser(writeCloser io.WriteCloser, remoteIdentity id.Identity) SecureWriteCloser {
	output := &secureWriteCloser{
		SourceField: NewSourceField(nil),
		WriteCloser: writeCloser,
		identity:    remoteIdentity,
	}
	return output
}

func (s *secureWriteCloser) Identity() id.Identity {
	return s.identity
}
