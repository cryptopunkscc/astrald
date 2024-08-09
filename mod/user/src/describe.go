package user

import (
	"context"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/lib/desc"
	"github.com/cryptopunkscc/astrald/mod/dir"
	"github.com/cryptopunkscc/astrald/mod/user"
)

var _ dir.Describer = &Module{}

func (mod *Module) Describe(ctx context.Context, identity *astral.Identity, opts *desc.Opts) []*desc.Desc {
	if identity.IsEqual(mod.UserID()) {
		return []*desc.Desc{{
			Source: mod.node.Identity(),
			Data:   user.UserDesc{Name: mod.Dir.DisplayName(identity)},
		}}
	}

	return nil
}
