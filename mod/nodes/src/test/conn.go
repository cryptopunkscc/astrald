package nodes

import (
	"io"

	"github.com/cryptopunkscc/astrald/mod/exonet"
	"github.com/cryptopunkscc/astrald/streams"
)

type Conn struct {
	io.ReadWriteCloser
	outbound       bool
	localEndpoint  *Endpoint
	remoteEndpoint *Endpoint
}

func PipeConn(localEndpoint, remoteEndpoint *Endpoint) (in *Conn, out *Conn) {
	left, right := streams.Pipe()
	out = &Conn{left, true, localEndpoint, remoteEndpoint}
	in = &Conn{right, false, remoteEndpoint, localEndpoint}
	return
}

var _ exonet.Conn = &Conn{}

func (t Conn) Outbound() bool {
	return t.outbound
}

func (t Conn) LocalEndpoint() exonet.Endpoint {
	return t.localEndpoint
}

func (t Conn) RemoteEndpoint() exonet.Endpoint {
	return t.remoteEndpoint
}
