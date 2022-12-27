package tor

import (
	"bytes"
	"encoding/base32"
	"errors"
	"fmt"
	"github.com/cryptopunkscc/astrald/cslq"
	"github.com/cryptopunkscc/astrald/infra"
	"strconv"
	"strings"
)

// Type check
var _ infra.Addr = Addr{}

const (
	addrVersion = 3
	addrLen     = 35
	onionSuffix = ".ONION"
	packPattern = "c [35]c s"
)

// Addr holds information about a Tor address
type Addr struct {
	digest []byte
	port   uint16
}

// Unpack converts a binary representation of the address to a struct
func Unpack(data []byte) (Addr, error) {
	var (
		err      error
		version  int
		keyBytes []byte
		port     uint16
		dec      = cslq.NewDecoder(bytes.NewReader(data))
	)

	err = dec.Decode(packPattern, &version, &keyBytes, &port)
	if err != nil {
		return Addr{}, err
	}

	if version != addrVersion {
		return Addr{}, errors.New("invalid version")
	}

	return Addr{
		digest: keyBytes,
		port:   port,
	}, nil
}

// Parse parses a string representation of a Tor address (both v2 and v3 are supported)
func Parse(s string) (Addr, error) {
	var (
		err      error
		port     = defaultListenPort
		hostPort = strings.SplitN(s, ":", 2)
	)

	if len(hostPort) > 1 {
		port, err = strconv.Atoi(hostPort[1])
		if err != nil {
			return Addr{}, fmt.Errorf("invalid address: %w", err)
		}
	}

	var b32data = strings.TrimSuffix(strings.ToUpper(hostPort[0]), onionSuffix)

	bytes, err := base32.StdEncoding.DecodeString(b32data)
	if err != nil {
		return Addr{}, fmt.Errorf("invalid address: %w", err)
	}

	if len(bytes) != addrLen {
		return Addr{}, errors.New("invalid length")
	}

	addr := Addr{
		digest: bytes,
		port:   uint16(port),
	}

	return addr, nil
}

// Network returns the name of the network the address belongs to
func (addr Addr) Network() string {
	return NetworkName
}

// String returns a human-readable representation of the address
func (addr Addr) String() string {
	if addr.IsZero() {
		return "none"
	}

	return fmt.Sprintf("%s.onion:%d", strings.ToLower(base32.StdEncoding.EncodeToString(addr.digest)), addr.port)
}

// Pack returns binary representation of the address
func (addr Addr) Pack() []byte {
	var b = &bytes.Buffer{}

	err := cslq.Encode(b, packPattern, addrVersion, addr.digest, addr.port)
	if err != nil {
		return nil
	}

	return b.Bytes()
}

// IsZero returns true if the address has zero-value
func (addr Addr) IsZero() bool {
	return (addr.digest == nil) || (len(addr.digest) == 0)
}
