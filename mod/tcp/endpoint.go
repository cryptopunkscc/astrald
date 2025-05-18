package tcp

import (
	"bytes"
	"errors"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/term"
	"github.com/cryptopunkscc/astrald/mod/exonet"
	"io"
	"net"
	"strconv"
)

var _ exonet.Endpoint = &Endpoint{}
var _ astral.Object = &Endpoint{}

// Endpoint is an astral.Object that holds information about a TCP endpoint, i.e. an IP address and a port.
// Supports JSON and text.
type Endpoint struct {
	IP   IP
	Port astral.Uint16
}

func (e *Endpoint) ObjectType() string {
	return "mod.tcp.endpoint"
}

func (e Endpoint) WriteTo(w io.Writer) (n int64, err error) {
	return astral.Struct(e).WriteTo(w)
}

func (e *Endpoint) ReadFrom(r io.Reader) (n int64, err error) {
	return astral.Struct(e).ReadFrom(r)
}

// exonet.Endpoint

func (e *Endpoint) Address() string {
	return net.JoinHostPort(e.IP.String(), strconv.Itoa(int(e.Port)))
}

func (e *Endpoint) Network() string {
	return "tcp"
}

func (e *Endpoint) Pack() []byte {
	var b = &bytes.Buffer{}
	if _, err := e.WriteTo(b); err != nil {
		return nil
	}
	return b.Bytes()
}

// Text marshaling

func (e Endpoint) MarshalText() (text []byte, err error) {
	return []byte(e.Address()), nil
}

func (e *Endpoint) UnmarshalText(text []byte) error {
	h, p, err := net.SplitHostPort(string(text))
	if err != nil {
		return err
	}

	ip, err := ParseIP(h)
	if err != nil {
		return err
	}

	port, err := strconv.Atoi(p)
	if err != nil {
		return err
	}

	// check if port fits in 16 bits
	if (port >> 16) > 0 {
		return errors.New("port out of range")
	}

	e.IP = ip
	e.Port = astral.Uint16(port)

	return nil
}

// ...

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

func init() {
	_ = astral.DefaultBlueprints.Add(&Endpoint{})

	term.SetTranslateFunc(func(o *Endpoint) astral.Object {
		return &term.ColorString{
			Color: term.HighlightColor,
			Text:  astral.String32(o.String()),
		}
	})
}
