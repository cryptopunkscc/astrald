package presence

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/mod/presence/proto"
	"github.com/cryptopunkscc/astrald/node/infra/drivers/inet"
	"net"
	"time"
)

type Ad struct {
	Identity  id.Identity
	Alias     string
	Endpoint  inet.Endpoint
	Timestamp time.Time
	Flags     int
	UDPAddr   *net.UDPAddr
}

func (ad *Ad) DiscoverFlag() bool {
	return ad.Flags&proto.FlagDiscover != 0
}
