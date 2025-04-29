package objects

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/mod/shell"
	"github.com/cryptopunkscc/astrald/object"
)

type opLoadArgs struct {
	ID   *object.ID
	Out  string      `query:"optional"`
	Zone astral.Zone `query:"optional"`
}

// OpLoad loads an object into memory and writes it to the output. OpLoad verifies the object hash.
func (mod *Module) OpLoad(ctx *astral.Context, q shell.Query, args opLoadArgs) (err error) {
	ctx = ctx.IncludeZone(args.Zone)

	object, err := objects.Load[astral.Object](ctx, mod.Root(), args.ID, mod.Blueprints())
	if err != nil {
		mod.log.Errorv(2, "error loading object: %v", err)
		return q.RejectWithCode(astral.CodeInternalError)
	}

	ch := astral.NewChannelFmt(q.Accept(), "", args.Out)
	defer ch.Close()

	return ch.Write(object)
}
