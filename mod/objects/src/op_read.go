package objects

import (
	"io"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/lib/routing"
	"github.com/cryptopunkscc/astrald/mod/auth"
	"github.com/cryptopunkscc/astrald/mod/objects"
)

type opReadArgs struct {
	ID     *astral.ObjectID
	Offset astral.Uint64 `query:"optional"`
	Limit  astral.Uint64 `query:"optional"`
	Zone   astral.Zone   `query:"optional"`
	Repo   string        `query:"optional"`
}

func (mod *Module) OpRead(ctx *astral.Context, q *routing.IncomingQuery, args opReadArgs) (err error) {
	ctx = ctx.IncludeZone(args.Zone)

	repo := mod.ReadDefault()

	mod.Auth.Authorize(ctx, &objects.ReadObjectAction{
		Action:   auth.NewAction(q.Caller()),
		ObjectID: args.ID,
	})

	if len(args.Repo) > 0 {
		repo = mod.GetRepository(args.Repo)
		if repo == nil {
			return q.Reject()
		}
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

	conn := q.AcceptRaw()
	defer conn.Close()

	_, err = io.Copy(conn, r)

	return err
}
