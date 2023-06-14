package net

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	"io"
)

var _ SecureWriteCloser = &secureWriteCloser{}

type secureWriteCloser struct {
	io.WriteCloser
	remoteIdentity id.Identity
}

func NewSecureWriteCloser(writeCloser io.WriteCloser, remoteIdentity id.Identity) SecureWriteCloser {
	return &secureWriteCloser{WriteCloser: writeCloser, remoteIdentity: remoteIdentity}
}

func (s *secureWriteCloser) RemoteIdentity() id.Identity {
	return s.remoteIdentity
}
