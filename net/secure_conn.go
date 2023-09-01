package net

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	"io"
)

var _ SecureConn = &secureConn{}
var _ WriterIter = &secureConn{}

type secureConn struct {
	SecureWriteCloser
	io.Reader
	localIdentity id.Identity
	outbound      bool
}

func NewSecureConn(localWriter SecureWriteCloser, localReader io.Reader, localIdentity id.Identity) SecureConn {
	return &secureConn{
		SecureWriteCloser: localWriter,
		Reader:            localReader,
		localIdentity:     localIdentity,
	}
}

func (s *secureConn) Read(p []byte) (n int, err error) {
	n, err = s.Reader.Read(p)
	if err != nil {
		s.Close()
	}
	return n, err
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
	return s.SecureWriteCloser.Identity()
}

func (s *secureConn) LocalIdentity() id.Identity {
	return s.localIdentity
}

func (s *secureConn) NextWriter() io.Writer {
	return s.SecureWriteCloser
}

func (s *secureConn) NextReader() io.Reader {
	return s.Reader
}
