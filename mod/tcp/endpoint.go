package tcp

import (
	"bytes"
	"errors"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/exonet"
	"github.com/cryptopunkscc/astrald/streams"
	"io"
	"net"
	"strconv"
)

var _ exonet.Endpoint = &Endpoint{}
var _ astral.Object = &Endpoint{}

type Endpoint struct {
	IP   IP
	Port astral.Uint16
}

func (e *Endpoint) ObjectType() string {
	return "astral.net.tcp.endpoint"
}

func (e *Endpoint) Address() string {
	return net.JoinHostPort(e.IP.String(), strconv.Itoa(int(e.Port)))
}

func (e *Endpoint) Network() string {
	return "tcp"
}

func (e Endpoint) WriteTo(w io.Writer) (n int64, err error) {
	return streams.WriteAllTo(w, e.IP, e.Port)
}

func (e *Endpoint) ReadFrom(r io.Reader) (n int64, err error) {
	return streams.ReadAllFrom(r, &e.IP, &e.Port)
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
	return e == nil || e.IP == nil
}

func ParseEndpoint(s string) (*Endpoint, error) {
	hostStr, portStr, err := net.SplitHostPort(s)
	if err != nil {
		return nil, err
	}

	ip, err := ParseIP(hostStr)
	if err != nil {
		return nil, err
	}

	port, err := strconv.Atoi(portStr)
	if err != nil {
		return nil, err
	}

	// check if port fits in 16 bits
	if (port >> 16) > 0 {
		return nil, errors.New("port out of range")
	}

	return &Endpoint{
		IP:   ip,
		Port: astral.Uint16(port),
	}, nil
}
