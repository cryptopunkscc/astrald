package user

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/shell"
	"github.com/cryptopunkscc/astrald/mod/user"
)

type opSwarmStatusArgs struct {
	In  string `query:"optional"`
	Out string `query:"optional"`
}

func (mod *Module) OpSwarmStatus(ctx *astral.Context, q shell.Query, args opSwarmStatusArgs) (err error) {
	ch := astral.NewChannelFmt(q.Accept(), args.In, args.Out)
	defer ch.Close()

	// list all nodes

	var swarmMap astral.Map
	swarm := mod.LocalSwarm()
	for _, node := range swarm {
		alias, err := mod.Dir.GetAlias(node)
		if err != nil {
			mod.log.Error("error getting alias for node %v: %v", node, err)
		}

		isLinked := mod.Nodes.IsLinked(node)

		// TODO: ask what else should be part of information about swarm member
		swarmMap.Set(node.String(), &user.SwarmMember{
			Alias:  astral.String(alias),
			Linked: astral.Bool(isLinked),
		})
	}

	return ch.Write(&swarmMap)
}
