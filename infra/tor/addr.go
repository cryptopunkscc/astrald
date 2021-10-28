package tor

import (
	bytes2 "bytes"
	"encoding/base32"
	"errors"
	"github.com/cryptopunkscc/astrald/infra"
	"io"
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
	packed := make([]byte, len(addr.bytes)+1)
	packed[0] = byte(addr.Version())
	copy(packed[1:], addr.bytes[:])
	return packed
}

func (addr Addr) IsZero() bool {
	return (addr.bytes == nil) || (len(addr.bytes) == 0)
}

func (addr Addr) Version() int {
	switch len(addr.bytes) {
	case 10:
		return 2
	case 35:
		return 3
	}
	return 0
}

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

func Unpack(data []byte) (Addr, error) {
	r := bytes2.NewReader(data)

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
