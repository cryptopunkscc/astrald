package objects

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/mod/shell"
)

type opScanArgs struct {
	Repo   string
	Follow bool         `query:"optional"`
	Zone   *astral.Zone `query:"optional"`
	Out    string       `query:"optional"`
}

// OpScan sends a list of object ids in a repository
func (mod *Module) OpScan(ctx *astral.Context, q shell.Query, args opScanArgs) (err error) {
	// prepare the context
	ctx = ctx.WithIdentity(q.Caller())
	if args.Zone == nil {
		ctx = ctx.WithZone(astral.ZoneAll)
	} else {
		ctx = ctx.WithZone(*args.Zone)
	}

	ctx, cancel := ctx.WithCancel()
	defer cancel()

	// accept
	ch := channel.New(q.Accept(), channel.WithOutputFormat(args.Out))
	defer ch.Close()

	// look up the repository
	repo := mod.GetRepository(args.Repo)
	if repo == nil {
		return ch.Send(astral.NewError("repository not found"))
	}

	// start the scan
	scan, err := repo.Scan(ctx, args.Follow)
	if err != nil {
		return ch.Send(astral.NewError(err.Error()))
	}

	// if the channel closes, cancel our scan context
	go func() {
		for {
			_, err := ch.Receive()
			if err != nil {
				cancel()
				return
			}
		}
	}()

	// forward scan results
	for id := range scan {
		err = ch.Send(id)
		if err != nil {
			return
		}
	}

	return ch.Send(&astral.EOS{})
}
