package tcp

import (
	"bytes"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/exonet"
	"github.com/cryptopunkscc/astrald/mod/tcp"
	"github.com/cryptopunkscc/astrald/streams"
	"io"
	"net"
	"strconv"
)

var _ exonet.Endpoint = &Endpoint{}
var _ astral.Object = &Endpoint{}

type Endpoint struct {
	ip   tcp.IP
	port astral.Uint16
}

func (e *Endpoint) ObjectType() string {
	return "astral.net.tcp.endpoint"
}

func (e *Endpoint) Address() string {
	return net.JoinHostPort(e.ip.String(), strconv.Itoa(int(e.port)))
}

func (e *Endpoint) Network() string {
	return "tcp"
}

func (e Endpoint) WriteTo(w io.Writer) (n int64, err error) {
	return streams.WriteAllTo(w, e.ip, e.port)
}

func (e *Endpoint) ReadFrom(r io.Reader) (n int64, err error) {
	return streams.ReadAllFrom(r, &e.ip, &e.port)
}

func (e *Endpoint) Pack() []byte {
	var b = &bytes.Buffer{}
	if _, err := e.WriteTo(b); err != nil {
		return nil
	}
	return b.Bytes()
}

func (e *Endpoint) String() string {
	return e.Address()
}

func (e *Endpoint) IsZero() bool {
	return e == nil || e.ip == nil
}
