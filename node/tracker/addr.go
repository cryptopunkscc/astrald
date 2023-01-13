package tracker

import (
	"github.com/cryptopunkscc/astrald/infra"
	"time"
)

type Addr struct {
	infra.Addr
	ExpiresAt time.Time
}

type dbAddr struct {
	NodeID    string
	Network   string
	Address   string
	ExpiresAt time.Time
}
