package contacts

import (
	"github.com/cryptopunkscc/astrald/infra"
	"time"
)

type Addr struct {
	infra.Addr
	ExpiresAt time.Time
}
