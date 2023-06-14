package net

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	"io"
)

var _ SecureConn = &secureConn{}

type secureConn struct {
	SecureWriteCloser
	io.Reader
	localIdentity id.Identity
	outbound      bool
}

func NewSecureConn(remoteWriter SecureWriteCloser, localReader io.Reader, localIdentity id.Identity) SecureConn {
	return &secureConn{
		SecureWriteCloser: remoteWriter,
		Reader:            localReader,
		localIdentity:     localIdentity,
	}
}

func (s *secureConn) Outbound() bool {
	return true
}

func (s *secureConn) LocalEndpoint() Endpoint {
	return nil
}

func (s *secureConn) RemoteEndpoint() Endpoint {
	return nil
}

func (s *secureConn) RemoteIdentity() id.Identity {
	return s.SecureWriteCloser.RemoteIdentity()
}

func (s *secureConn) LocalIdentity() id.Identity {
	return s.localIdentity
}
