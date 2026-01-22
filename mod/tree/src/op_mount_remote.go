package tree

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/mod/shell"
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

	err = mod.MountRemote(ctx, args.Path, targetID, args.Root)
	if err != nil {
		return ch.Send(astral.Err(err))
	}

	return ch.Send(&astral.Ack{})
}
