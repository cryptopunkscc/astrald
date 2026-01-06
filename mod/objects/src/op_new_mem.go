package objects

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/mod/objects/mem"
	"github.com/cryptopunkscc/astrald/mod/shell"
)

type opNewMemArgs struct {
	Name string
	Size string `query:"optional"`
	In   string `query:"optional"`
	Out  string `query:"optional"`
}

func (mod *Module) OpNewMem(ctx *astral.Context, q shell.Query, args opNewMemArgs) (err error) {
	ch := channel.New(q.Accept(), channel.WithFormats(args.In, args.Out))
	defer ch.Close()

	// parse the size
	size := astral.Size(mem.DefaultSize)
	if len(args.Size) > 0 {
		size, err = astral.ParseSize(args.Size)
		if err != nil {
			return ch.Send(astral.NewError(err.Error()))
		}
	}

	// create the repository
	repo := mem.New("Memory ("+args.Name+")", int64(size))
	err = mod.AddRepository(args.Name, repo)
	if err != nil {
		return ch.Send(astral.NewError(err.Error()))
	}

	mod.AddGroup(objects.RepoMemory, args.Name)

	return ch.Send(&astral.Ack{})
}
