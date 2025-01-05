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
}

func (mod *Module) OpRead(ctx astral.Context, q shell.Query, args opReadArgs) (err error) {
	if !mod.Auth.Authorize(q.Caller(), objects.ActionRead, &args.ID) {
		return q.Reject()
	}

	if q.Origin() == "network" {
		args.Zone &= ^astral.ZoneVirtual
	}

	if args.Zone == 0 {
		args.Zone = astral.ZoneDevice
	}

	r, err := mod.Open(ctx, args.ID, &objects.OpenOpts{
		Zone:   args.Zone,
		Offset: uint64(args.Offset),
	})
	if err != nil {
		mod.log.Errorv(2, "open %v error: %v", args.ID, err)
		return q.Reject()
	}

	conn, err := q.Accept()
	if err != nil {
		return err
	}
	defer conn.Close()

	_, err = io.Copy(conn, r)

	return err
}
