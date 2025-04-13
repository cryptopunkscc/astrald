package kos

import (
	"encoding/json"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/shell"
)

type opGetArgs struct {
	Key    string
	Format string `query:"optional"`
}

func (mod *Module) OpGet(ctx *astral.Context, q shell.Query, args opGetArgs) error {
	typ, payload, err := mod.db.Get(ctx.Identity(), args.Key)
	if err != nil {
		return q.Reject()
	}

	conn, err := q.Accept()
	if err != nil {
		return err
	}
	defer conn.Close()

	raw := &astral.RawObject{Type: typ, Payload: payload}

	switch args.Format {
	case "", "bin":
		_, err = astral.Write(conn, raw, false)

	case "json":
		var obj astral.Object
		obj, err = mod.Objects.Blueprints().Refine(raw)
		if err == nil {
			err = json.NewEncoder(conn).Encode(map[string]interface{}{
				"type": obj.ObjectType(),
				"data": obj,
			})
		} else {
			err = json.NewEncoder(conn).Encode(map[string]interface{}{
				"type":    raw.Type,
				"payload": raw.Payload,
			})
		}
	}

	return err
}
