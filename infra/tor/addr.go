package tor

import (
	"bytes"
	"encoding/base32"
	"errors"
	"fmt"
	"github.com/cryptopunkscc/astrald/cslq"
	"github.com/cryptopunkscc/astrald/infra"
	"io"
	"strconv"
	"strings"
)

const v2KeyLength = 10
const v3KeyLength = 35

// Type check
var _ infra.Addr = Addr{}

// Addr holds information about a Tor address
type Addr struct {
	digest []byte
	port   uint16
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
	b := &bytes.Buffer{}
	cslq.Write(b, byte(addr.Version()))
	cslq.Write(b, addr.digest)
	cslq.Write(b, addr.port)
	return b.Bytes()
}

// IsZero returns true if the address has zero-value
func (addr Addr) IsZero() bool {
	return (addr.digest == nil) || (len(addr.digest) == 0)
}

// Version returns the version of Tor address (2 or 3) or 0 if the address data is errorous
func (addr Addr) Version() int {
	switch len(addr.digest) {
	case v2KeyLength:
		return 2
	case v3KeyLength:
		return 3
	}
	return 0
}

// Parse parses a string representation of a Tor address (both v2 and v3 are supported)
func Parse(s string) (Addr, error) {
	var err error
	var port = defaultListenPort
	var hostPort = strings.SplitN(s, ":", 2)

	if len(hostPort) > 1 {
		port, err = strconv.Atoi(hostPort[1])
		if err != nil {
			return Addr{}, fmt.Errorf("invalid address: %w", err)
		}
	}

	b32data := strings.TrimSuffix(strings.ToUpper(hostPort[0]), ".ONION")

	bytes, err := base32.StdEncoding.DecodeString(b32data)
	if err != nil {
		return Addr{}, fmt.Errorf("invalid address: %w", err)
	}

	addr := Addr{
		digest: bytes,
		port:   uint16(port),
	}

	if addr.Version() == 0 {
		return Addr{}, errors.New("invalid address")
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
		keyBytes = make([]byte, v2KeyLength)
	case 3:
		keyBytes = make([]byte, v3KeyLength)
	default:
		return Addr{}, errors.New("invalid version")
	}

	_, err = io.ReadFull(r, keyBytes)
	if err != nil {
		return Addr{}, err
	}

	port, err := cslq.ReadUint16(r)
	if err != nil {
		return Addr{}, errors.New("invalid port")
	}

	return Addr{
		digest: keyBytes,
		port:   port,
	}, nil
}
