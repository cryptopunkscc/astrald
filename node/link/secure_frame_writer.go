package link

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/mux"
	"github.com/cryptopunkscc/astrald/net"
)

var _ net.SecureWriteCloser = &SecureFrameWriter{}

type SecureFrameWriter struct {
	*mux.FrameWriter
	remoteID id.Identity
	link     *Link
}

func NewSecureFrameWriter(l *Link, remotePort int) *SecureFrameWriter {
	return &SecureFrameWriter{
		FrameWriter: mux.NewFrameWriter(l.mux, remotePort),
		remoteID:    l.RemoteIdentity(),
		link:        l,
	}
}

func (s *SecureFrameWriter) RemoteIdentity() id.Identity {
	return s.remoteID
}

func (s *SecureFrameWriter) Link() *Link {
	return s.link
}
