package nat

import (
	"bytes"
	"fmt"
	"io"
	"net"
	"strconv"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/ip"
)

// NOTE: this is same for UDP and TCP - consider moving to a common package

// UDPEndpoint is an astral.Object that holds information about a UDP endpoint,
type UDPEndpoint struct {
	IP   ip.IP
	Port astral.Uint16
}

func (e UDPEndpoint) ObjectType() string {
	return "mod.nat.udp_endpoint"
}

func (e UDPEndpoint) WriteTo(w io.Writer) (n int64, err error) {
	return astral.Objectify(&e).WriteTo(w)
}

func (e *UDPEndpoint) ReadFrom(r io.Reader) (n int64, err error) {
	return astral.Objectify(e).ReadFrom(r)
}

func (e *UDPEndpoint) Address() string {
	return net.JoinHostPort(e.IP.String(), strconv.Itoa(int(e.Port)))
}

func (e *UDPEndpoint) Network() string {
	return "utp"
}

func (e *UDPEndpoint) HostString() string {
	return e.IP.String()
}

func (e *UDPEndpoint) PortNumber() int {
	return int(e.Port)
}

func (e *UDPEndpoint) Pack() []byte {
	var b = &bytes.Buffer{}
	if _, err := e.WriteTo(b); err != nil {
		return nil
	}
	return b.Bytes()
}

func (e UDPEndpoint) MarshalText() (text []byte, err error) {
	return []byte(e.Address()), nil
}

func (e *UDPEndpoint) UnmarshalText(text []byte) error {
	h, p, err := net.SplitHostPort(string(text))
	if err != nil {
		return err
	}

	ip, err := ip.ParseIP(h)
	if err != nil {
		return err
	}

	port, err := strconv.Atoi(p)
	if err != nil {
		return err
	}

	// check if port fits in 16 bits
	if (port >> 16) > 0 {
		return fmt.Errorf("port out of range")
	}

	e.IP = ip
	e.Port = astral.Uint16(port)

	return nil
}

func (e *UDPEndpoint) String() string {
	return e.Address()
}

func (e *UDPEndpoint) IsZero() bool {
	return e == nil || e.IP == nil
}

func ParseEndpoint(s string) (*UDPEndpoint, error) {
	hostStr, portStr, err := net.SplitHostPort(s)
	if err != nil {
		return nil, err
	}

	ip, err := ip.ParseIP(hostStr)
	if err != nil {
		return nil, err
	}

	port, err := strconv.Atoi(portStr)
	if err != nil {
		return nil, err
	}

	if (port >> 16) > 0 {
		return nil, fmt.Errorf("port out of range")
	}

	return &UDPEndpoint{
		IP:   ip,
		Port: astral.Uint16(port),
	}, nil
}

func (e *UDPEndpoint) UDPAddr() *net.UDPAddr {
	return &net.UDPAddr{
		IP:   net.ParseIP(e.IP.String()),
		Port: int(e.Port),
	}
}

func init() {
	_ = astral.Add(&UDPEndpoint{})
}
