package user

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/shell"
	"github.com/cryptopunkscc/astrald/mod/user"
)

type opInfoArgs struct {
	Format string `query:"optional"`
}

func (mod *Module) OpInfo(ctx *astral.Context, q shell.Query, args opInfoArgs) (err error) {
	ac := mod.ActiveContract()
	if ac == nil {
		return q.Reject()
	}

	if !q.Caller().IsEqual(ac.UserID) {
		return q.Reject()
	}

	ch := astral.NewChannel(q.Accept(), args.Format)
	defer ch.Close()

	return ch.Write(&user.Info{
		NodeAlias: astral.String8(mod.Dir.DisplayName(ac.NodeID)),
		UserAlias: astral.String8(mod.Dir.DisplayName(ac.UserID)),
	})
}
