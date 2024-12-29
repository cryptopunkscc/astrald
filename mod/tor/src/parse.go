package tor

import (
	"encoding/base32"
	"errors"
	"fmt"
	"github.com/cryptopunkscc/astrald/mod/exonet"
	"strconv"
	"strings"
)

var _ exonet.Parser = &Module{}

func (mod *Module) Parse(network string, address string) (exonet.Endpoint, error) {
	return Parse(address)
}

// Parse parses a string representation of a Driver address (both log and v3 are supported)
func Parse(s string) (*Endpoint, error) {
	var (
		err      error
		port     = defaultListenPort
		hostPort = strings.SplitN(s, ":", 2)
	)

	if len(hostPort) > 1 {
		port, err = strconv.Atoi(hostPort[1])
		if err != nil {
			return nil, fmt.Errorf("invalid address: %w", err)
		}
	}

	var b32data = strings.TrimSuffix(strings.ToUpper(hostPort[0]), onionSuffix)

	bytes, err := base32.StdEncoding.DecodeString(b32data)
	if err != nil {
		return nil, fmt.Errorf("invalid address: %w", err)
	}

	if len(bytes) != addrLen {
		return nil, errors.New("invalid length")
	}

	e := &Endpoint{
		digest: bytes,
		port:   uint16(port),
	}

	return e, nil
}
