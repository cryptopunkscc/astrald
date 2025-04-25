package tor

import (
	"bytes"
	"errors"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/cslq"
	"github.com/cryptopunkscc/astrald/mod/exonet"
	"github.com/cryptopunkscc/astrald/mod/tor"
)

const addrVersion = 3
const packPattern = "c [35]c s"

var _ exonet.Unpacker = &Module{}

func (mod *Module) Unpack(network string, data []byte) (exonet.Endpoint, error) {
	if network != tor.ModuleName {
		return nil, exonet.ErrUnsupportedNetwork
	}
	return Unpack(data)
}

// Unpack converts a binary representation of the address to a struct
func Unpack(data []byte) (*tor.Endpoint, error) {
	var (
		err      error
		version  int
		keyBytes []byte
		port     uint16
		dec      = cslq.NewDecoder(bytes.NewReader(data))
	)

	err = dec.Decodef(packPattern, &version, &keyBytes, &port)
	if err != nil {
		return nil, err
	}

	if version != addrVersion {
		return nil, errors.New("invalid version")
	}

	return &tor.Endpoint{
		Digest: keyBytes,
		Port:   astral.Uint16(port),
	}, nil
}
