package user

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/ops"
	"github.com/cryptopunkscc/astrald/mod/auth"
	"github.com/cryptopunkscc/astrald/mod/user"
)

type opSwarmStatusArgs struct {
	In  string `query:"optional"`
	Out string `query:"optional"`
}

func (mod *Module) OpSwarmStatus(ctx *astral.Context, q *ops.Query, args opSwarmStatusArgs) (err error) {
	ac := mod.ActiveContract()
	if ac == nil {
		return q.RejectWithCode(2)
	}

	ch := channel.New(q.Accept(), channel.WithFormats(args.In, args.Out))
	defer ch.Close()

	contracts, err := mod.ActiveContractsOf(ac.Issuer)
	if err != nil {
		return ch.Send(astral.NewError(err.Error()))
	}

	byNodeID := make(map[string]*auth.SignedContract)
	for _, c := range contracts {
		byNodeID[c.Subject.String()] = c
	}

	swarm := mod.LocalSwarm()
	for _, node := range swarm {
		alias, aliasErr := mod.Dir.GetAlias(node)
		if aliasErr != nil {
			mod.log.Error("error getting alias for node %v: %v", node, aliasErr)
		}

		sc, ok := byNodeID[node.String()]
		if !ok {
			mod.log.Error("no active contract found for node %v", node)
			continue
		}

		contractID, err := astral.ResolveObjectID(sc)
		if err != nil {
			return ch.Send(astral.NewError(err.Error()))
		}

		err = ch.Send(&user.SwarmMember{
			SignedContractID: contractID,
			Identity:         node,
			Alias:            astral.String8(alias),
			Linked:           astral.Bool(mod.Nodes.IsLinked(node)),
			Contract:         sc.Contract,
		})
		if err != nil {
			return ch.Send(astral.NewError(err.Error()))
		}
	}

	return ch.Send(&astral.EOS{})
}
