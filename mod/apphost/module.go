package apphost

import (
	"github.com/cryptopunkscc/astrald/astral"
)

const ModuleName = "apphost"
const DBPrefix = "apphost__"

type Module interface {
	SetDefaultIdentity(*astral.Identity) error
	DefaultIdentity() *astral.Identity
	CreateAccessToken(identity *astral.Identity) (string, error)
}
