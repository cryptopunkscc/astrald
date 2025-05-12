package tor

import (
	"bytes"
	"errors"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/exonet"
	"github.com/cryptopunkscc/astrald/mod/tor"
	"io"
)

const addrVersion = 3

var _ exonet.Unpacker = &Module{}

func (mod *Module) Unpack(network string, data []byte) (exonet.Endpoint, error) {
	if network != tor.ModuleName {
		return nil, exonet.ErrUnsupportedNetwork
	}
	return Unpack(data)
}

// Unpack converts a binary representation of the address to a struct
func Unpack(data []byte) (*tor.Endpoint, error) {
	r := bytes.NewReader(data)

	var (
		err      error
		version  uint8
		keyBytes = make([]byte, 35)
		port     uint16
	)

	_, err = (*astral.Uint8)(&version).ReadFrom(r)
	if err != nil {
		return nil, err
	}

	if version != addrVersion {
		return nil, errors.New("invalid version")
	}

	_, err = io.ReadFull(r, keyBytes)
	if err != nil {
		return nil, err
	}

	_, err = (*astral.Uint16)(&port).ReadFrom(r)
	if err != nil {
		return nil, err
	}

	return &tor.Endpoint{
		Digest: keyBytes,
		Port:   astral.Uint16(port),
	}, nil
}
