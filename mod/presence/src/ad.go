package presence

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/mod/presence/proto"
	"github.com/cryptopunkscc/astrald/net"
	_net "net"
	"time"
)

type Ad struct {
	Identity  id.Identity
	Alias     string
	Endpoint  net.Endpoint
	Timestamp time.Time
	Flags     int
	UDPAddr   *_net.UDPAddr
}

func (ad *Ad) DiscoverFlag() bool {
	return ad.Flags&proto.FlagDiscover != 0
}
