package tcp

import (
	"errors"
	"github.com/cryptopunkscc/astrald/astral"
	"io"
	"net"
)

type IP net.IP

func (IP) ObjectType() string {
	return "astral.net.ip.address"
}

func (ip IP) WriteTo(w io.Writer) (n int64, err error) {
	var ver = 0
	var p = ip[:]
	switch len(p) {
	case net.IPv4len:
		ver = 0
	case net.IPv6len:
		if ip4 := []byte(net.IP(ip).To4()); ip4 != nil {
			p = ip4
		} else {
			ver = 1
		}
	default:
		return 0, errors.New("invalid IP length")
	}

	var m int
	n, err = astral.Uint8(ver).WriteTo(w)
	if err != nil {
		return
	}
	m, err = w.Write(p)
	n += int64(m)

	return
}

func (ip *IP) ReadFrom(r io.Reader) (n int64, err error) {
	var v astral.Uint8
	n, err = v.ReadFrom(r)
	if err != nil {
		return
	}

	var l int

	switch v {
	case 0:
		l = net.IPv4len
	case 1:
		l = net.IPv6len
	default:
		err = errors.New("unknown IP version")
		return
	}

	var m int
	var buf = make([]byte, l)
	m, err = io.ReadFull(r, buf)
	n += int64(m)
	if err != nil {
		return
	}
	*ip = IP(buf[:m])
	return
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

func ParseIP(s string) (IP, error) {
	return IP(net.ParseIP(s)), nil
}
