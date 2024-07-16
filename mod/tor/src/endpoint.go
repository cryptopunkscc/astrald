package tor

import (
	"bytes"
	"encoding/base32"
	"fmt"
	"github.com/cryptopunkscc/astrald/cslq"
	"github.com/cryptopunkscc/astrald/mod/exonet"
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

// Endpoint holds information about a Driver address
type Endpoint struct {
	digest []byte
	port   uint16
}

// Network returns the name of the network the address belongs to
func (addr *Endpoint) Network() string {
	return ModuleName
}

// Address returns a human-readable representation of the address
func (addr *Endpoint) Address() string {
	if addr.IsZero() {
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
	return (addr.digest == nil) || (len(addr.digest) == 0)
}
