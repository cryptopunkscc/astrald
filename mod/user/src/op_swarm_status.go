package user

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/mod/shell"
	"github.com/cryptopunkscc/astrald/mod/user"
)

type opSwarmStatusArgs struct {
	In  string `query:"optional"`
	Out string `query:"optional"`
}

func (mod *Module) OpSwarmStatus(ctx *astral.Context, q shell.Query, args opSwarmStatusArgs) (err error) {
	ac := mod.ActiveContract()
	if ac == nil {
		return q.RejectWithCode(2)
	}

	ch := channel.New(q.Accept(), channel.WithFormats(args.In, args.Out))
	defer ch.Close()

	contracts, err := mod.ActiveContractsOf(ac.UserID)
	if err != nil {
		return ch.Send(astral.NewError(err.Error()))
	}

	contractsByNodeID := make(map[string]*user.SignedNodeContract)
	for _, c := range contracts {
		contractsByNodeID[c.NodeID.String()] = c
	}

	swarm := mod.LocalSwarm()
	for _, node := range swarm {
		alias, err := mod.Dir.GetAlias(node)
		if err != nil {
			mod.log.Error("error getting alias for node %v: %v", node, err)
		}

		contract, ok := contractsByNodeID[node.String()]
		if !ok {
			mod.log.Error("no active contract found for node %v", node)
		}

		contractID, err := astral.ResolveObjectID(contract)
		if err != nil {
			return ch.Send(astral.NewError(err.Error()))
		}

		err = ch.Send(&user.SwarmMember{
			SignedContractID: contractID,
			Identity:         node,
			Alias:            astral.String8(alias),
			Linked:           astral.Bool(mod.Nodes.IsLinked(node)),
			Contract:         contract.NodeContract,
		})
		if err != nil {
			return ch.Send(astral.NewError(err.Error()))
		}
	}

	return ch.Send(&astral.EOS{})
}
