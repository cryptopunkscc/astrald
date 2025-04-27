package objects

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/mod/shell"
	"github.com/cryptopunkscc/astrald/object"
	"io"
)

type opReadArgs struct {
	ID     object.ID
	Offset astral.Uint64 `query:"optional"`
	Zone   astral.Zone   `query:"optional"`
	Size   astral.Uint64 `query:"optional"`
}

func (mod *Module) OpRead(ctx *astral.Context, q shell.Query, args opReadArgs) (err error) {
	if q.Origin() == "network" {
		args.Zone &= ^astral.ZoneVirtual
	}

	ctx = ctx.IncludeZone(args.Zone).WithIdentity(q.Caller())

	r, err := mod.Open(ctx, args.ID, &objects.OpenOpts{
		Offset: uint64(args.Offset),
	})
	if err != nil {
		mod.log.Errorv(2, "open %v error: %v", args.ID, err)
		return q.Reject()
	}
	defer r.Close()

	conn := q.Accept()
	defer conn.Close()

	if args.Size > 0 {
		_, err = io.CopyN(conn, r, int64(args.Size))
	} else {
		_, err = io.Copy(conn, r)
	}

	return err
}
