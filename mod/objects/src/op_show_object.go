package objects

import (
	"encoding/json"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/mod/shell"
	"github.com/cryptopunkscc/astrald/object"
)

// objects.show_object
//
// Returns an object in JSON.

type opShowObjectArgs struct {
	ID object.ID
}

func (mod *Module) OpShowObject(ctx *astral.Context, q shell.Query, args opShowObjectArgs) (err error) {
	if !mod.Auth.Authorize(q.Caller(), objects.ActionRead, &args.ID) {
		return q.Reject()
	}

	obj, err := mod.Load(args.ID)
	if err != nil {
		mod.log.Errorv(2, "get %v error: %v", args.ID, err)
		return q.Reject()
	}

	conn, err := q.Accept()
	if err != nil {
		return err
	}
	defer conn.Close()

	return json.NewEncoder(conn).Encode(obj)
}
