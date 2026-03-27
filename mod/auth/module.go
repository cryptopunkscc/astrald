package auth

import (
	"github.com/cryptopunkscc/astrald/astral"
)

const ModuleName = "auth"

type Module interface {
	Authorize(ctx *astral.Context, identity *astral.Identity, action Action, target astral.Object) bool
	AddAuthorizer(action Action, handlers ...Handler)
}

const ActionSudo = "mod.admin.sudo"
