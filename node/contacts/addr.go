package contacts

import (
	"github.com/cryptopunkscc/astrald/infra"
	"time"
)

type Addr struct {
	infra.Addr
	ExpiresAt time.Time
}

func wrapAddr(addr infra.Addr) *Addr {
	return &Addr{
		Addr:      addr,
		ExpiresAt: time.Now().Add(defaultAddressValidity),
	}
}
