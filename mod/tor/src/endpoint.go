package tor

import (
	"bytes"
	"encoding/base32"
	"fmt"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/term"
	"github.com/cryptopunkscc/astrald/cslq"
	"github.com/cryptopunkscc/astrald/mod/exonet"
	"github.com/cryptopunkscc/astrald/streams"
	"io"
	"strings"
)

// Type check
var _ exonet.Endpoint = &Endpoint{}

const (
	addrVersion = 3
	addrLen     = 35
	onionSuffix = ".ONION"
	packPattern = "c [35]c s"
)

var _ astral.Object = &Endpoint{}

// Endpoint holds information about a Driver address
type Endpoint struct {
	digest []byte
	port   uint16
}

func (addr *Endpoint) ObjectType() string {
	return "astrald.mod.tor.endpoint"
}

// Network returns the name of the network the address belongs to
func (addr *Endpoint) Network() string {
	return ModuleName
}

// Address returns a human-readable representation of the address
func (addr *Endpoint) Address() string {
	if addr == nil || addr.IsZero() {
		return "none"
	}

	return fmt.Sprintf("%s.onion:%d", strings.ToLower(base32.StdEncoding.EncodeToString(addr.digest)), addr.port)
}

// Pack returns binary representation of the address
func (addr *Endpoint) Pack() []byte {
	var b = &bytes.Buffer{}

	err := cslq.Encode(b, packPattern, addrVersion, addr.digest, addr.port)
	if err != nil {
		return nil
	}

	return b.Bytes()
}

// IsZero returns true if the address has zero-value
func (addr *Endpoint) IsZero() bool {
	return (addr == nil) || (addr.digest == nil) || (len(addr.digest) == 0)
}

func (addr *Endpoint) String() string {
	return addr.Network() + ":" + addr.Address()
}

func (addr *Endpoint) WriteTo(w io.Writer) (n int64, err error) {
	return streams.WriteAllTo(w,
		astral.Bytes8(addr.digest),
		astral.Uint16(addr.port),
	)
}

func (addr *Endpoint) ReadFrom(r io.Reader) (n int64, err error) {
	return streams.ReadAllFrom(r,
		(*astral.Bytes8)(&addr.digest),
		(*astral.Uint16)(&addr.port),
	)
}

func init() {
	term.SetTranslateFunc(func(o *Endpoint) astral.Object {
		return &term.ColorString{
			Color: term.HighlightColor,
			Text:  astral.String32(o.String()),
		}
	})
}
