package user

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/lib/routing"
	"github.com/cryptopunkscc/astrald/mod/user"
)

type opSwarmStatusArgs struct {
	In  string `query:"optional"`
	Out string `query:"optional"`
}

func (mod *Module) OpSwarmStatus(ctx *astral.Context, q *routing.IncomingQuery, args opSwarmStatusArgs) (err error) {
	ac := mod.ActiveContract()
	if ac == nil {
		return q.RejectWithCode(2)
	}

	ch := q.Accept()
	defer ch.Close()

	for _, node := range mod.ActiveNodes(ac.Issuer) {
		alias, aliasErr := mod.Dir.GetAlias(node)
		if aliasErr != nil {
			mod.log.Error("error getting alias for node %v: %v", node, aliasErr)
		}

		err = ch.Send(&user.SwarmMember{
			Identity: node,
			Alias:    astral.String8(alias),
			Linked:   astral.Bool(mod.Nodes.IsLinked(node)),
		})
		if err != nil {
			return ch.Send(astral.NewError(err.Error()))
		}
	}

	return ch.Send(&astral.EOS{})
}
