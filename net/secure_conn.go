package net

import (
	"github.com/cryptopunkscc/astrald/auth/id"
)

var _ SecureConn = &secureConn{}
var _ SourceGetter = &secureConn{}
var _ OutputGetter = &secureConn{}

type secureConn struct {
	SecureWriteCloser
	SecureReader
	outbound bool
}

func NewSecureConn(w SecureWriteCloser, r SecureReader, outbound bool) SecureConn {
	return &secureConn{
		SecureWriteCloser: NewConnOutput(w),
		SecureReader:      r,
		outbound:          outbound,
	}
}

func (s *secureConn) Read(p []byte) (n int, err error) {
	n, err = s.SecureReader.Read(p)
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
	return s.SecureReader.Identity()
}

func (s *secureConn) Source() any {
	return s.SecureReader
}

func (s *secureConn) Output() SecureWriteCloser {
	return s.SecureWriteCloser
}

type ConnOutput struct {
	*OutputField
}

func NewConnOutput(output SecureWriteCloser) *ConnOutput {
	w := &ConnOutput{}
	w.OutputField = NewOutputField(w, output)
	return w
}

func (out *ConnOutput) Identity() id.Identity {
	return out.Output().Identity()
}

func (out *ConnOutput) Write(p []byte) (n int, err error) {
	return out.Output().Write(p)
}

func (out *ConnOutput) Close() error {
	return out.Output().Close()
}
