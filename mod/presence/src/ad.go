package presence

import (
	"github.com/cryptopunkscc/astrald/id"
	"github.com/cryptopunkscc/astrald/mod/exonet"
	_net "net"
	"slices"
	"time"
)

type Ad struct {
	Identity  id.Identity
	Alias     string
	Endpoint  exonet.Endpoint
	ExpiresAt time.Time
	Flags     []string
	UDPAddr   *_net.UDPAddr
}

func (ad *Ad) Has(flag string) bool {
	return slices.Contains(ad.Flags, flag)
}
