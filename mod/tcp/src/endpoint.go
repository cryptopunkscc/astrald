package tcp

import (
	"bytes"
	"github.com/cryptopunkscc/astrald/cslq"
	"github.com/cryptopunkscc/astrald/net"
	_net "net"
	"strconv"
)

const (
	ipv4 = iota // IPv4 (32 bit)
	ipv6        // IPv6 (128 bit)
)

var _ net.Endpoint = Endpoint{}

type Endpoint struct {
	ver  int
	ip   _net.IP
	port uint16
}

func (e Endpoint) Pack() []byte {
	var b = &bytes.Buffer{}

	switch e.ver {
	case ipv4:
		cslq.Encode(b, "x00 [4]c s", e.ip[len(e.ip)-4:], e.port)
	case ipv6:
		cslq.Encode(b, "x01 [16]c s", e.ip, e.port)
	}

	return b.Bytes()
}

func (e Endpoint) String() string {
	ip := e.ip.String()

	if e.ver == ipv6 {
		ip = "[" + ip + "]"
	}

	if e.port != 0 {
		ip = ip + ":" + strconv.Itoa(int(e.port))
	}
	return ip
}

func (e Endpoint) Network() string {
	return "tcp"
}

func (e Endpoint) Port() int {
	return int(e.port)
}

func (e Endpoint) IP() _net.IP {
	return e.ip
}

func (e Endpoint) IsGlobalUnicast() bool {
	return e.ip.IsGlobalUnicast()
}

// IsPrivate returns true if the endpoint belongs to a private network (like LAN)
func (e Endpoint) IsPrivate() bool {
	return e.ip.IsPrivate()
}

func (e Endpoint) IsPublicUnicast() bool {
	return !e.IsPrivate() && e.IsGlobalUnicast()
}

func (e Endpoint) IsZero() bool {
	if e.ip == nil {
		return true
	}
	return false
}
