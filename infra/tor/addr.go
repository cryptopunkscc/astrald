package tor

import (
	"encoding/base32"
	"errors"
	"github.com/cryptopunkscc/astrald/infra"
	"strings"
)

var _ infra.Addr = Addr{}

type Addr struct {
	bytes []byte
}

func (addr Addr) Network() string {
	return NetworkName
}

func (addr Addr) String() string {
	if len(addr.bytes) == 0 {
		return ""
	}
	return strings.ToLower(base32.StdEncoding.EncodeToString(addr.bytes)) + ".onion"
}

func (addr Addr) Pack() []byte {
	packed := make([]byte, len(addr.bytes))
	copy(packed[:], addr.bytes[:])
	return packed
}

func (addr Addr) IsZero() bool {
	return (addr.bytes == nil) || (len(addr.bytes) == 0)
}

func Parse(s string) (Addr, error) {
	b32data := strings.TrimSuffix(strings.ToUpper(s), ".ONION")

	if len(b32data) != 56 {
		return Addr{}, errors.New("not a valid tor v3 address")
	}

	bytes, err := base32.StdEncoding.DecodeString(b32data)
	if err != nil {
		return Addr{}, err
	}

	return Addr{bytes: bytes}, nil
}

func Unpack(addr []byte) (Addr, error) {
	if len(addr) != 35 {
		return Addr{}, errors.New("invalid data size")
	}
	bytes := make([]byte, 35)
	copy(bytes[:], addr[:])
	return Addr{
		bytes: bytes,
	}, nil
}
