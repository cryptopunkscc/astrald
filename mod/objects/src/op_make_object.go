package objects

import (
	"encoding/json"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/shell"
)

type opMakeObjectArgs struct {
	Type      string
	Format    string `query:"optional"`
	Canonical bool   `query:"optional"`
}

func (mod *Module) OpMakeObject(ctx *astral.Context, q shell.Query, args opMakeObjectArgs) (err error) {
	object := mod.Blueprints().Make(args.Type)
	if object == nil {
		mod.log.Errorv(2, "objects.make_object: unknown type %v", args.Type)
		return q.Reject()
	}

	ch := astral.NewChannel(q.Accept(), args.Format)
	defer ch.Close()

	err = json.NewDecoder(ch.Transport()).Decode(object)
	if err != nil {
		var e = astral.NewError(err.Error())
		if args.Canonical {
			_, err = astral.WriteCanonical(ch.Transport(), e)
			return
		}
		return ch.Write(e)
	}

	if args.Canonical {
		_, err = astral.WriteCanonical(ch.Transport(), object)
		return
	}
	return ch.Write(object)
}
