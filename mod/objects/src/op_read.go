package objects

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/shell"
	"github.com/cryptopunkscc/astrald/object"
	"io"
)

type opReadArgs struct {
	ID     *object.ID
	Offset astral.Uint64 `query:"optional"`
	Limit  astral.Uint64 `query:"optional"`
	Zone   astral.Zone   `query:"optional"`
	Repo   astral.String `query:"optional"`
}

func (mod *Module) OpRead(ctx *astral.Context, q shell.Query, args opReadArgs) (err error) {
	ctx = ctx.IncludeZone(args.Zone)

	repo, err := mod.GetRepository(args.Repo.String())
	if err != nil {
		return
	}

	r, err := repo.Read(
		ctx.WithIdentity(q.Caller()),
		args.ID,
		int64(args.Offset),
		int64(args.Limit),
	)
	if err != nil {
		mod.log.Errorv(2, "read %v error: %v", args.ID, err)
		return q.Reject()
	}
	defer r.Close()

	conn := q.Accept()
	defer conn.Close()

	_, err = io.Copy(conn, r)

	return err
}
