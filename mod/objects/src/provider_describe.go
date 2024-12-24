package objects

import (
	"context"
	"encoding/json"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/lib/query"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/object"
	"io"
	"time"
)

type describeArgs struct {
	ID     object.ID
	Format string `query:"optional"`
	Zones  string `query:"optional"`
}

type jsonDescriptor struct {
	Type string
	Data any
}

func (p *Provider) Describe(ctx context.Context, q *astral.Query, w io.WriteCloser) (io.WriteCloser, error) {
	var args describeArgs
	_, err := query.ParseTo(q.Query, &args)
	if err != nil {
		p.mod.log.Errorv(1, "describe: error parsing query: %v", err)
		return query.Reject()
	}

	if !p.mod.Auth.Authorize(q.Caller, objects.ActionRead, &args.ID) {
		return query.Reject()
	}

	scope := astral.DefaultScope()

	if len(args.Zones) > 0 {
		scope.Zone = astral.Zones(args.Zones)
	}

	return query.Accept(q, w, func(conn astral.Conn) {
		defer conn.Close()

		ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
		defer cancel()

		descs, err := p.mod.DescribeObject(ctx, args.ID, scope)
		if err != nil {
			return
		}

		for so := range descs {
			if !p.mod.Auth.Authorize(q.Caller, objects.ActionReadDescriptor, so.Object) {
				continue
			}

			switch args.Format {
			case "json":
				err = json.NewEncoder(conn).Encode(jsonDescriptor{
					Type: so.Object.ObjectType(),
					Data: so.Object,
				})
				if err != nil {
					p.mod.log.Errorv(1, "describe.json: error encoding json: %v", err)
					return
				}

			case "":
				_, err = astral.Write(conn, so.Object, false)
				if err != nil {
					p.mod.log.Errorv(1, "describe: error writing object: %v", err)
					return
				}
			}
		}
	})
}
