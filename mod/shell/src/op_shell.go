package shell

import (
	"errors"
	"io"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/auth"
	"github.com/cryptopunkscc/astrald/mod/shell"
)

type opShellArgs struct {
	As astral.String8 `query:"optional"`
}

func (mod *Module) OpShell(ctx *astral.Context, query shell.Query, args opShellArgs) (err error) {
	// handle args
	if len(args.As) > 0 {
		asID, err := mod.Dir.ResolveIdentity(string(args.As))
		if err != nil {
			return err
		}

		if !mod.Auth.Authorize(query.Caller(), auth.ActionSudo, asID) {
			return astral.NewError("access denied")
		}

		ctx = ctx.WithIdentity(asID)
	} else {
		ctx = ctx.WithIdentity(query.Caller())
	}

	// accept
	var conn io.ReadWriteCloser
	conn = query.Accept()
	defer conn.Close()

	// handle session
	err = NewSession(mod, conn).Run(ctx)
	switch {
	case err == nil, errors.Is(err, io.EOF):
		mod.log.Logv(1, "shell session with %v ended", ctx.Identity())
		err = nil
	default:
		mod.log.Errorv(1, "shell session with %v ended in error: %v", ctx.Identity(), err)
	}

	return
}
