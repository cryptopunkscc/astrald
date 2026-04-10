package auth

import "github.com/cryptopunkscc/astrald/astral"

func (mod *Module) Ban(ctx *astral.Context, identity *astral.Identity) error {
	return mod.db.addBan(identity)
}

func (mod *Module) IsBanned(identity *astral.Identity) bool {
	return mod.db.isBanned(identity)
}
