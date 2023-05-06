package bt

import (
	"fmt"
	"github.com/cryptopunkscc/astrald/net"
)

var _ net.Endpoint = &Endpoint{}

type Endpoint struct {
	mac [6]byte
}

func (e Endpoint) Network() string {
	return DriverName
}

func (e Endpoint) String() string {
	return fmt.Sprintf("%2.2X:%2.2X:%2.2X:%2.2X:%2.2X:%2.2X",
		e.mac[5], e.mac[4], e.mac[3], e.mac[2], e.mac[1], e.mac[0])
}

func (e Endpoint) Pack() []byte {
	buf := make([]byte, 6)
	copy(buf, e.mac[:])
	return buf
}

func (e Endpoint) IsZero() bool {
	for _, b := range e.mac {
		if b != 0 {
			return false
		}
	}
	return true
}
