package gw

import (
	"bytes"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/cslq"
	"github.com/cryptopunkscc/astrald/infra"
	"strings"
)

var _ infra.Addr = Addr{}

type Addr struct {
	gate   id.Identity
	cookie string
}

// NewAddr insatntiates and returns a new Addr
func NewAddr(gate id.Identity, cookie string) Addr {
	return Addr{gate: gate, cookie: cookie}
}

// Unpack converts a binary representation of the address to a struct
func Unpack(data []byte) (Addr, error) {
	r := bytes.NewReader(data)

	var (
		nodeID id.Identity
		cookie string
	)

	err := cslq.Decode(r, "v [c]c", &nodeID, &cookie)
	if err != nil {
		return Addr{}, err
	}

	return Addr{
		gate:   nodeID,
		cookie: cookie,
	}, nil
}

// Parse converts a text representation of a gateway address to an Addr struct
func Parse(str string) (Addr, error) {
	parts := strings.SplitN(str, ":", 2)

	gate, err := id.ParsePublicKeyHex(parts[0])
	if err != nil {
		return Addr{}, err
	}

	var cookie string
	if len(parts) == 2 {
		cookie = parts[1]
	}

	return Addr{
		gate:   gate,
		cookie: cookie,
	}, nil
}

// Pack returns a binary representation of the address
func (a Addr) Pack() []byte {
	buf := &bytes.Buffer{}

	cslq.Encode(buf, "v [c]c", a.gate, a.cookie)

	return buf.Bytes()
}

// String returns a text representation of the address
func (a Addr) String() string {
	if a.IsZero() {
		return "unknown"
	}
	return a.gate.PublicKeyHex() + ":" + a.cookie
}

func (a Addr) IsZero() bool {
	return a.gate.IsZero()
}

func (a Addr) Cookie() string {
	return a.cookie
}

func (a Addr) Gate() id.Identity {
	return a.gate
}

func (a Addr) Network() string {
	return NetworkName
}
