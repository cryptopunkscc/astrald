package query

import (
	"github.com/cryptopunkscc/astrald/astral"
	"io"
)

// conn is a basic implementation of an astral.Conn over a Reader and a WriteCloser.
type conn struct {
	localID  *astral.Identity
	remoteID *astral.Identity
	io.WriteCloser
	io.Reader
	outbound bool
}

var _ astral.Conn = &conn{}

func newConn(localID *astral.Identity, remoteID *astral.Identity, w io.WriteCloser, r io.Reader, outbound bool) astral.Conn {
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

func (s *conn) RemoteIdentity() *astral.Identity {
	return s.remoteID
}

func (s *conn) LocalIdentity() *astral.Identity {
	return s.localID
}
