package tree

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/astrald"
	"github.com/cryptopunkscc/astrald/mod/shell"
	"github.com/cryptopunkscc/astrald/mod/tree"
)

type opMountRemoteArgs struct {
	Path   string
	Target string
	Root   string `query:"optional"`
	In     string `query:"optional"`
	Out    string `query:"optional"`
}

func (mod *Module) OpMountRemote(ctx *astral.Context, q shell.Query, args opMountRemoteArgs) (err error) {
	ch := channel.New(q.Accept(), channel.WithFormats(args.In, args.Out))
	defer ch.Close()

	targetID, err := mod.Dir.ResolveIdentity(args.Target)
	if err != nil {
		return ch.Send(astral.NewError(err.Error()))
	}

	treeClient := astrald.NewTreeClient(astrald.DefaultClient(), targetID)

	root := treeClient.Root()

	if len(args.Root) > 0 {
		root, err = tree.Query(ctx, root, args.Root, false)
		if err != nil {
			return ch.Send(astral.NewError(err.Error()))
		}
	}

	err = mod.Mount(args.Path, root)
	if err != nil {
		return ch.Send(astral.NewError(err.Error()))
	}

	return ch.Send(&astral.Ack{})
}
