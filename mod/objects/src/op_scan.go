package objects

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/shell"
)

type opScanArgs struct {
	Type string      `query:"optional"`
	Repo string      `query:"optional"`
	Out  string      `query:"optional"`
	Zone astral.Zone `query:"optional"`
}

// OpScan sends a list of object ids in a repository
func (mod *Module) OpScan(ctx *astral.Context, q shell.Query, args opScanArgs) (err error) {
	ctx = ctx.IncludeZone(args.Zone)

	repo, err := mod.GetRepository(args.Repo)
	if err != nil || repo == nil {
		return q.Reject()
	}

	scanCh, err := repo.Scan(ctx.WithIdentity(q.Caller()))
	if err != nil {
		return q.Reject()
	}

	ch := astral.NewChannelFmt(q.Accept(), "", args.Out)
	defer ch.Close()

	for id := range scanCh {
		if len(args.Type) > 0 {
			t, err := mod.GetType(ctx, id)
			if err != nil {
				continue
			}
			if args.Type != t {
				continue
			}
		}

		err = ch.Write(id)
		if err != nil {
			return
		}
	}

	return
}
