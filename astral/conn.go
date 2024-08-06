package astral

import (
	"io"
)

// Conn defines the basic interface of an astral connection
type Conn interface {
	Read(b []byte) (n int, err error)
	Write(b []byte) (n int, err error)
	Close() error
	LocalIdentity() *Identity
	RemoteIdentity() *Identity
}

var _ Conn = &conn{}

type conn struct {
	localID  *Identity
	remoteID *Identity
	io.WriteCloser
	io.Reader
	outbound bool
}

func newConn(localID *Identity, remoteID *Identity, w io.WriteCloser, r io.Reader, outbound bool) Conn {
	return &conn{
		localID:     localID,
		remoteID:    remoteID,
		WriteCloser: w,
		Reader:      r,
		outbound:    outbound,
	}
}

func (s *conn) Read(p []byte) (n int, err error) {
	n, err = s.Reader.Read(p)
	if err != nil {
		s.Close()
	}
	return n, err
}

func (s *conn) Outbound() bool {
	return s.outbound
}

func (s *conn) RemoteIdentity() *Identity {
	return s.remoteID
}

func (s *conn) LocalIdentity() *Identity {
	return s.localID
}
