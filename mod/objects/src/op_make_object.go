package objects

import (
	"encoding/json"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/shell"
)

type opMakeObjectArgs struct {
	Type string
}

func (mod *Module) OpMakeObject(ctx *astral.Context, q shell.Query, args opMakeObjectArgs) (err error) {
	object := mod.Blueprints().Make(args.Type)
	if object == nil {
		mod.log.Errorv(2, "objects.make_object: unknown type %v", args.Type)
		return q.Reject()
	}

	conn := q.Accept()
	defer conn.Close()

	err = json.NewDecoder(conn).Decode(&object)

	if err != nil {
		astral.WriteCanonical(conn, astral.NewError(err.Error()))
		return
	}

	astral.WriteCanonical(conn, object)

	return
}
