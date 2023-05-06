package bt

import (
	"errors"
	"github.com/cryptopunkscc/astrald/net"
)

func (drv *Driver) Unpack(network string, data []byte) (net.Endpoint, error) {
	if network != DriverName {
		return nil, errors.New("invalid network")
	}
	return Unpack(data)
}

func Unpack(addr []byte) (Endpoint, error) {
	if len(addr) != 6 {
		return Endpoint{}, errors.New("invalid data length")
	}
	var a Endpoint
	copy(a.mac[:], addr)
	return a, nil
}
