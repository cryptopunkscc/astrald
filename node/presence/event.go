package presence

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/infra"
)

type EventIdentityPresent struct {
	Identity id.Identity
	Addr     infra.Addr
}

type EventIdentityGone struct {
	Identity id.Identity
}
