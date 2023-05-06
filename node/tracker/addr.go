package tracker

import (
	"github.com/cryptopunkscc/astrald/net"
	"time"
)

type Addr struct {
	net.Endpoint
	ExpiresAt time.Time
}

type dbAddr struct {
	NodeID    string
	Network   string
	Address   string
	ExpiresAt time.Time
}
