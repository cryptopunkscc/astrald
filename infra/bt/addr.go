package bt

import (
	"errors"
	"fmt"
	"github.com/cryptopunkscc/astrald/infra"
	"strconv"
	"strings"
)

var _ infra.Addr = &Addr{}

type Addr struct {
	mac [6]byte
}

func (a Addr) Network() string {
	return NetworkName
}

func (a Addr) String() string {
	return fmt.Sprintf("%2.2X:%2.2X:%2.2X:%2.2X:%2.2X:%2.2X",
		a.mac[5], a.mac[4], a.mac[3], a.mac[2], a.mac[1], a.mac[0])
}

func (a Addr) Pack() []byte {
	buf := make([]byte, 6)
	copy(buf, a.mac[:])
	return buf
}

func (a Addr) IsZero() bool {
	for _, b := range a.mac {
		if b != 0 {
			return false
		}
	}
	return true
}

func Unpack(addr []byte) (Addr, error) {
	if len(addr) != 6 {
		return Addr{}, errors.New("invalid data length")
	}
	var a Addr
	copy(a.mac[:], addr)
	return a, nil
}

func Parse(addr string) (Addr, error) {
	a := strings.Split(addr, ":")
	if len(a) != 6 {
		return Addr{}, infra.ErrInvalidAddress
	}

	var mac [6]byte

	for i, b := range a {
		u, err := strconv.ParseUint(b, 16, 8)
		if err != nil {
			return Addr{}, infra.ErrInvalidAddress
		}
		mac[len(mac)-1-i] = byte(u)
	}

	return Addr{mac: mac}, nil
}
