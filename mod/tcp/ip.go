package tcp

import (
	"encoding/json"
	"errors"
	"github.com/cryptopunkscc/astrald/astral"
	"io"
	"net"
)

type IP net.IP

func (IP) ObjectType() string {
	return "mod.tcp.ip_address"
}

func (ip IP) WriteTo(w io.Writer) (n int64, err error) {
	return astral.Bytes8(ip).WriteTo(w)
}

func (ip *IP) ReadFrom(r io.Reader) (n int64, err error) {
	return (*astral.Bytes8)(ip).ReadFrom(r)
}

func (ip IP) String() string {
	return net.IP(ip).String()
}

func (ip IP) IsIPv4() bool {
	return net.IP(ip).To4() != nil
}

func (ip IP) IsIPv6() bool {
	return net.IP(ip).To16() != nil
}

func (ip IP) IsLoopback() bool { return net.IP(ip).IsLoopback() }

func (ip IP) IsGlobalUnicast() bool { return net.IP(ip).IsGlobalUnicast() }

func (ip IP) IsPrivate() bool { return net.IP(ip).IsPrivate() }

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

func ParseIP(s string) (IP, error) {
	return IP(net.ParseIP(s)), nil
}
