package fs

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/ops"
	"github.com/cryptopunkscc/astrald/mod/objects"
)

type opNewWatchArgs struct {
	Path  string
	Name  string
	Label string `query:"optional"`
	In    string `query:"optional"`
	Out   string `query:"optional"`
}

func (mod *Module) OpNewWatch(ctx *astral.Context, q *ops.Query, args opNewWatchArgs) (err error) {
	ch := q.AcceptChannel(channel.WithFormats(args.In, args.Out))
	defer ch.Close()

	if args.Label == "" {
		args.Label = args.Name
	}

	repo, err := NewWatchRepository(mod, args.Path, args.Label)
	if err != nil {
		return ch.Send(astral.Err(err))
	}

	scanCtx, cancel := mod.ctx.WithCancel()
	repo.scanCancel = cancel

	go func() {
		if err := mod.indexer.scan(scanCtx, args.Path, true); err != nil {
			mod.log.Error("scan %v: %v", args.Path, err)
		}
	}()

	err = mod.Objects.AddRepository(args.Name, repo)
	if err != nil {
		cancel()
		return ch.Send(astral.Err(err))
	}

	err = mod.Objects.AddGroup(objects.RepoLocal, args.Name)
	if err != nil {
		cancel()
		return ch.Send(astral.Err(err))
	}

	return ch.Send(&astral.Ack{})
}
