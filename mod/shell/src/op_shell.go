package shell

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/admin"
	"github.com/cryptopunkscc/astrald/mod/shell"
	"io"
)

type opShellArgs struct {
	As astral.String `query:"optional"`
}

func (mod *Module) opShell(ctx astral.Context, query shell.Query, args opShellArgs) (err error) {
	// handle args
	if len(args.As) > 0 {
		asID, err := mod.Dir.ResolveIdentity(string(args.As))
		if err != nil {
			return err
		}

		if !mod.Auth.Authorize(query.Caller(), admin.ActionSudo, asID) {
			return astral.NewError("access denied")
		}

		ctx = astral.WrapContext(ctx, asID)
	}

	// accept
	var conn io.ReadWriteCloser
	conn, err = query.Accept()
	if err != nil {
		return err
	}
	defer conn.Close()

	// handle session
	err = NewSession(mod, conn).Run(ctx)
	if err != nil {
		mod.log.Errorv(2, "session with %v ended in error: %v", ctx.Identity(), err)
	}

	return
}
