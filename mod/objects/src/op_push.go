package objects

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/channel"
	"github.com/cryptopunkscc/astrald/lib/ops"
)

const maxPushSize = 32 * 1024

type opPushArgs struct {
	In  string `query:"optional"`
	Out string `query:"optional"`
}

func (mod *Module) OpPush(ctx *astral.Context, q *ops.Query, args opPushArgs) (err error) {
	ch := channel.New(q.Accept(), channel.WithFormats(args.In, args.Out))
	defer ch.Close()

	return ch.Collect(func(o astral.Object) (err error) {
		objectID, _ := astral.ResolveObjectID(o)

		var ok = astral.Bool(mod.receive(q.Caller(), o))
		if ok {
			mod.log.Logv(1, "received %v (%v) from %v", o.ObjectType(), objectID, q.Caller())
		} else {
			mod.log.Logv(1, "rejected %v (%v) from %v", o.ObjectType(), objectID, q.Caller())
		}
		err = ch.Send(&ok)
		return
	})
}
