package tor

import (
	"bytes"
	"encoding/base32"
	"errors"
	"github.com/cryptopunkscc/astrald/infra"
	"io"
	"strings"
)

// Type check
var _ infra.Addr = Addr{}

// Addr holds information about a Tor address
type Addr struct {
	bytes []byte
}

// Network returns the name of the network the address belongs to
func (addr Addr) Network() string {
	return NetworkName
}

// String returns a human-readable representation of the address
func (addr Addr) String() string {
	if len(addr.bytes) == 0 {
		return "unknown"
	}
	return strings.ToLower(base32.StdEncoding.EncodeToString(addr.bytes)) + ".onion"
}

// Pack returns binary representation of the address
func (addr Addr) Pack() []byte {
	packed := make([]byte, len(addr.bytes)+1)
	packed[0] = byte(addr.Version())
	copy(packed[1:], addr.bytes[:])
	return packed
}

// IsZero returns true if the address has zero-value
func (addr Addr) IsZero() bool {
	return (addr.bytes == nil) || (len(addr.bytes) == 0)
}

// Version returns the version of Tor address (2 or 3) or 0 if the address data is errorous
func (addr Addr) Version() int {
	switch len(addr.bytes) {
	case 10:
		return 2
	case 35:
		return 3
	}
	return 0
}

// Parse parses a string representation of a Tor address (both v2 and v3 are supported)
func Parse(s string) (Addr, error) {
	b32data := strings.TrimSuffix(strings.ToUpper(s), ".ONION")

	bytes, err := base32.StdEncoding.DecodeString(b32data)
	if err != nil {
		return Addr{}, err
	}

	addr := Addr{bytes: bytes}
	if addr.Version() == 0 {
		return Addr{}, errors.New("not a supported tor address")
	}

	return addr, nil
}

// Unpack converts a binary representation of the address to a struct
func Unpack(data []byte) (Addr, error) {
	r := bytes.NewReader(data)

	version, err := r.ReadByte()
	if err != nil {
		return Addr{}, err
	}

	var keyBytes []byte

	switch version {
	case 2:
		keyBytes = make([]byte, 10)
	case 3:
		keyBytes = make([]byte, 35)
	default:
		return Addr{}, errors.New("invalid version")
	}

	_, err = io.ReadFull(r, keyBytes)
	if err != nil {
		return Addr{}, err
	}

	return Addr{
		bytes: keyBytes,
	}, nil
}
