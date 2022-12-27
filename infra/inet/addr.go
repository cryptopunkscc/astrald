package inet

import (
	"bytes"
	"errors"
	"github.com/cryptopunkscc/astrald/cslq"
	"github.com/cryptopunkscc/astrald/infra"
	"net"
	"strconv"
)

const NetworkName = "inet"

const (
	ipv4 = iota // IPv4 (32 bit)
	ipv6        // IPv6 (128 bit)
)

var _ infra.Addr = Addr{}

type Addr struct {
	ver  int
	ip   net.IP
	port uint16
}

func Unpack(buf []byte) (addr *Addr, err error) {
	addr = &Addr{}

	var r = bytes.NewReader(buf)

	if err = cslq.Decode(r, "c", &addr.ver); err != nil {
		return
	}

	switch addr.ver {
	case ipv4:
		return addr, cslq.Decode(r, "[4]c s", &addr.ip, &addr.port)
	case ipv6:
		return addr, cslq.Decode(r, "[16]c s", &addr.ip, &addr.port)
	}

	return addr, errors.New("invalid version")
}

func Parse(s string) (addr Addr, err error) {
	var host, port string

	host, port, err = net.SplitHostPort(s)
	if err != nil {
		return
	}

	addr.ip = net.ParseIP(host)
	if addr.ip == nil {
		return addr, errors.New("invalid ip")
	}

	if addr.ip.To4() == nil {
		addr.ver = ipv6
	}

	var p int
	if p, err = strconv.Atoi(port); err != nil {
		return
	} else {
		if (p < 0) || (p > 65535) {
			return addr, errors.New("port out of range")
		}
		addr.port = uint16(p)
	}

	return
}

func (addr Addr) Pack() []byte {
	var b = &bytes.Buffer{}

	switch addr.ver {
	case ipv4:
		cslq.Encode(b, "x00 [4]c s", addr.ip[len(addr.ip)-4:], addr.port)
	case ipv6:
		cslq.Encode(b, "x01 [16]c s", addr.ip, addr.port)
	}

	return b.Bytes()
}

func (addr Addr) String() string {
	ip := addr.ip.String()

	if addr.ver == ipv6 {
		ip = "[" + ip + "]"
	}

	if addr.port != 0 {
		ip = ip + ":" + strconv.Itoa(int(addr.port))
	}
	return ip
}

func (addr Addr) Network() string {
	return NetworkName
}
