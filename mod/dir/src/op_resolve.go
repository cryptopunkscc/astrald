package dir

import (
	"encoding/json"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/shell"
)

type opResolveArgs struct {
	Name   astral.String
	Format astral.String `query:"optional"`
}

func (mod *Module) OpResolve(ctx *astral.Context, q shell.Query, args opResolveArgs) (err error) {
	if len(args.Name) == 0 {
		return q.Reject()
	}

	id, err := mod.ResolveIdentity(string(args.Name))
	if err != nil {
		return q.Reject()
	}

	conn, err := q.Accept()
	if err != nil {
		return err
	}
	defer conn.Close()

	switch args.Format {
	case "json":
		json.NewEncoder(conn).Encode(id)

	default:
		_, err = id.WriteTo(conn)
	}

	return
}
