package user

import (
	"context"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/lib/desc"
	"github.com/cryptopunkscc/astrald/mod/user"
)

func (mod *Module) Describe(ctx context.Context, identity id.Identity, opts *desc.Opts) []*desc.Desc {
	localUser := mod.LocalUser()
	if localUser == nil {
		return nil
	}

	if identity.IsEqual(localUser.Identity()) {
		return []*desc.Desc{{
			Source: mod.node.Identity(),
			Data:   user.UserDesc{Name: mod.node.Resolver().DisplayName(identity)},
		}}
	}

	return nil
}
