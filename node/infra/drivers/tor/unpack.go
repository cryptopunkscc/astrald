package tor

import (
	"bytes"
	"errors"
	"github.com/cryptopunkscc/astrald/cslq"
	"github.com/cryptopunkscc/astrald/net"
	"github.com/cryptopunkscc/astrald/node/infra"
)

var _ infra.Unpacker = &Driver{}

func (drv *Driver) Unpack(network string, data []byte) (net.Endpoint, error) {
	if network != DriverName {
		return nil, errors.New("invalid network")
	}
	return Unpack(data)
}

// Unpack converts a binary representation of the address to a struct
func Unpack(data []byte) (Endpoint, error) {
	var (
		err      error
		version  int
		keyBytes []byte
		port     uint16
		dec      = cslq.NewDecoder(bytes.NewReader(data))
	)

	err = dec.Decode(packPattern, &version, &keyBytes, &port)
	if err != nil {
		return Endpoint{}, err
	}

	if version != addrVersion {
		return Endpoint{}, errors.New("invalid version")
	}

	return Endpoint{
		digest: keyBytes,
		port:   port,
	}, nil
}
