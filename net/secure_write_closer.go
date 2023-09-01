package net

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	"io"
)

type SecureWriteCloser interface {
	io.WriteCloser
	Identity() id.Identity // Returns the remote identity
}

var _ SecureWriteCloser = &secureWriteCloser{}
var _ WriterIter = &secureWriteCloser{}

type secureWriteCloser struct {
	io.WriteCloser
	identity id.Identity
}

func NewSecureWriteCloser(writeCloser io.WriteCloser, remoteIdentity id.Identity) SecureWriteCloser {
	return &secureWriteCloser{WriteCloser: writeCloser, identity: remoteIdentity}
}

func (s *secureWriteCloser) Identity() id.Identity {
	return s.identity
}

func (s *secureWriteCloser) NextWriter() io.Writer {
	return s.WriteCloser
}
