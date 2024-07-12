package tcp

import (
	"errors"
	"github.com/cryptopunkscc/astrald/core"
	"github.com/cryptopunkscc/astrald/net"
	_net "net"
	"strconv"
)

func (mod *Module) Parse(network string, address string) (net.Endpoint, error) {
	switch network {
	case "tcp", "inet":
	default:
		return nil, core.ErrUnsupportedNetwork
	}

	return Parse(address)
}

func Parse(s string) (endpoint Endpoint, err error) {
	var host, port string

	host, port, err = _net.SplitHostPort(s)
	if err != nil {
		return
	}

	endpoint.ip = _net.ParseIP(host)
	if endpoint.ip == nil {
		return endpoint, errors.New("invalid ip")
	}

	if endpoint.ip.To4() == nil {
		endpoint.ver = ipv6
	}

	var p int
	if p, err = strconv.Atoi(port); err != nil {
		return
	} else {
		if (p < 0) || (p > 65535) {
			return endpoint, errors.New("port out of range")
		}
		endpoint.port = uint16(p)
	}

	return
}
