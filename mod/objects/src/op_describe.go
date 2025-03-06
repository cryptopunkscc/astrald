package objects

import (
	"context"
	"encoding/json"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/mod/shell"
	"github.com/cryptopunkscc/astrald/object"
	"time"
)

type opDescribeArgs struct {
	ID     object.ID
	Format string      `query:"optional"`
	Zones  astral.Zone `query:"optional"`
}

type jsonDescriptor struct {
	Type string
	Data any
}

func (mod *Module) OpDescribe(ctx astral.Context, q shell.Query, args opDescribeArgs) (err error) {
	if !mod.Auth.Authorize(q.Caller(), objects.ActionRead, &args.ID) {
		return q.Reject()
	}

	if q.Origin() == "network" {
		args.Zones &= ^astral.ZoneVirtual
	}

	scope := astral.DefaultScope()
	scope.Zone = args.Zones

	stream, err := shell.AcceptStream(q)
	if err != nil {
		return
	}
	defer stream.Close()

	tctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	descs, err := mod.Describe(tctx, args.ID, scope)
	if err != nil {
		return
	}

	for so := range descs {
		if !mod.Auth.Authorize(q.Caller(), objects.ActionReadDescriptor, so.Object) {
			continue
		}

		switch args.Format {
		case "json":
			err = json.NewEncoder(stream).Encode(jsonDescriptor{
				Type: so.Object.ObjectType(),
				Data: so.Object,
			})
			if err != nil {
				mod.log.Errorv(1, "describe.json: error encoding json: %v", err)
				return
			}

		case "":
			_, err = stream.WriteObject(so.Object)
			if err != nil {
				mod.log.Errorv(1, "describe: error writing object: %v", err)
				return
			}
		}
	}

	return
}
