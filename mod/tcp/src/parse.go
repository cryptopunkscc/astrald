package tcp

import (
	"errors"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/exonet"
	"github.com/cryptopunkscc/astrald/mod/tcp"
	"net"
	"strconv"
)

func (mod *Module) Parse(network string, address string) (exonet.Endpoint, error) {
	switch network {
	case "tcp", "inet":
	default:
		return nil, exonet.ErrUnsupportedNetwork
	}

	return Parse(address)
}

func Parse(s string) (*Endpoint, error) {
	hostStr, portStr, err := net.SplitHostPort(s)
	if err != nil {
		return nil, err
	}

	ip, err := tcp.ParseIP(hostStr)
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
		ip:   ip,
		port: astral.Uint16(port),
	}, nil
}
