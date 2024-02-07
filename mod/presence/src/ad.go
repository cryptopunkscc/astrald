package presence

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/mod/presence"
	"github.com/cryptopunkscc/astrald/net"
	_net "net"
	"slices"
	"time"
)

type Ad struct {
	Identity  id.Identity
	Alias     string
	Endpoint  net.Endpoint
	Timestamp time.Time
	Flags     []string
	UDPAddr   *_net.UDPAddr
}

func (ad *Ad) DiscoverFlag() bool {
	return slices.Contains(ad.Flags, presence.DiscoverFlag)
}
