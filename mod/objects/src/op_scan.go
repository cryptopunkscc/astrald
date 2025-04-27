package objects

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/shell"
)

type opScanArgs struct {
	Repo   string `query:"optional"`
	Format string `query:"optional"`
}

func (mod *Module) OpScan(ctx *astral.Context, q shell.Query, args opScanArgs) (err error) {
	repo, err := mod.GetRepository(args.Repo)
	if err != nil || repo == nil {
		return q.Reject()
	}

	scanCh, err := repo.Scan(ctx.WithIdentity(q.Caller()))
	if err != nil {
		return q.Reject()
	}

	ch := astral.NewChannel(q.Accept(), args.Format)
	defer ch.Close()

	for id := range scanCh {
		err = ch.Write(id)
		if err != nil {
			return
		}
	}

	return
}
