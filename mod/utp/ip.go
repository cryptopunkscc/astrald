package utp

import (
	"encoding/json"
	"errors"
	"io"
	"net"

	"github.com/cryptopunkscc/astrald/astral"
)

// NOTE: this is same for tcp / udp modules, consider moving to a common package
type IP net.IP

func ParseIP(s string) (IP, error) {
	return IP(net.ParseIP(s)), nil
}

// astral

func (IP) ObjectType() string {
	return "mod.utp.ip_address"
}

func (ip IP) WriteTo(w io.Writer) (n int64, err error) {
	if ip.IsIPv4() {
		return astral.Bytes8(net.IP(ip).To4()).WriteTo(w)
	}

	return astral.Bytes8(ip).WriteTo(w)
}

func (ip *IP) ReadFrom(r io.Reader) (n int64, err error) {
	return (*astral.Bytes8)(ip).ReadFrom(r)
}

// json

func (ip IP) MarshalJSON() ([]byte, error) {
	return json.Marshal(ip.String())
}

func (ip *IP) UnmarshalJSON(b []byte) error {
	var str string
	err := json.Unmarshal(b, &str)
	if err != nil {
		return nil
	}

	parsed := IP(net.ParseIP(str))

	if parsed == nil {
		return errors.New("invalid IP")
	}

	*ip = parsed
	return nil
}

// text

func (ip IP) MarshalText() (text []byte, err error) {
	return []byte(ip.String()), nil
}

func (ip *IP) UnmarshalText(text []byte) error {
	parsed := IP(net.ParseIP(string(text)))
	if parsed == nil {
		return errors.New("invalid IP")
	}
	*ip = parsed
	return nil
}

// ...

func (ip IP) IsIPv4() bool {
	return net.IP(ip).To4() != nil
}

func (ip IP) IsIPv6() bool {
	return net.IP(ip).To16() != nil
}

func (ip IP) IsLoopback() bool { return net.IP(ip).IsLoopback() }

func (ip IP) IsGlobalUnicast() bool { return net.IP(ip).IsGlobalUnicast() }

func (ip IP) IsPrivate() bool { return net.IP(ip).IsPrivate() }

func (ip IP) String() string {
	return net.IP(ip).String()
}

func init() {
	_ = astral.DefaultBlueprints.Add(&IP{})
}
