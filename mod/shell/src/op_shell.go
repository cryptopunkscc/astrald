package shell

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/admin"
	"github.com/cryptopunkscc/astrald/mod/shell"
)

type opShellArgs struct {
	As astral.String `query:"optional"`
}

func (mod *Module) opShell(ctx astral.Context, env *shell.Env, args opShellArgs) (err error) {
	if len(args.As) > 0 {
		asID, err := mod.Dir.ResolveIdentity(string(args.As))
		if err != nil {
			return err
		}

		if !mod.Auth.Authorize(ctx.Identitiy(), admin.ActionSudo, asID) {
			return astral.NewError("access denied")
		}

		ctx = astral.WrapContext(ctx, asID)
	}

	var s = NewSession(mod, env)
	err = s.Run(ctx)
	if err != nil {
		mod.log.Errorv(2, "session with %v ended in error: %v", ctx.Identitiy(), err)
	}
	return
}
