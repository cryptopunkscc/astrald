package apphost

import (
	"github.com/cryptopunkscc/astrald/astral"
)

type Guest struct {
	Token    string
	Identity *astral.Identity
	Endpoint string
}
