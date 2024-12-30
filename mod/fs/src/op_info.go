package fs

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/term"
	"github.com/cryptopunkscc/astrald/mod/fs"
	"github.com/cryptopunkscc/astrald/mod/shell"
	"slices"
)

func (mod *Module) opInfo(ctx astral.Context, env *shell.Env) error {
	if !mod.Auth.Authorize(ctx.Identitiy(), fs.ActionManage, nil) {
		return astral.NewError("unauthorized")
	}

	paths := mod.watcher.List()
	slices.Sort(paths)

	for _, path := range paths {
		var p = fs.Path(path)
		env.WriteObject(&p)
		env.WriteObject(&term.Newline{})
	}
	return nil
}
