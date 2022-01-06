package bt

import (
	"github.com/cryptopunkscc/astrald/infra"
	"golang.org/x/sys/unix"
)

var _ infra.Conn = &Conn{}

type Conn struct {
	nfd        int
	outbount   bool
	localAddr  Addr
	remoteAddr Addr
}

func (c Conn) Read(p []byte) (n int, err error) {
	return unix.Read(c.nfd, p)
}

func (c Conn) Write(p []byte) (n int, err error) {
	return unix.Write(c.nfd, p)
}

func (c Conn) Close() error {
	return unix.Close(c.nfd)
}

func (c Conn) Outbound() bool {
	return c.outbount
}

func (c Conn) LocalAddr() infra.Addr {
	return c.localAddr
}

func (c Conn) RemoteAddr() infra.Addr {
	return c.remoteAddr
}
