package astrald

import "github.com/cryptopunkscc/astrald/astral"

func NewContext() *astral.Context {
	return astral.
		NewContext(nil).
		WithIdentity(GuestID()).
		WithZone(astral.ZoneAll)
}
