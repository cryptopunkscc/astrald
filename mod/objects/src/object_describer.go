package objects

import (
	"context"
	"encoding/json"
	"errors"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/lib/query"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/object"
	"io"
	"slices"
)

type describeArgs struct {
	ID     object.ID
	Format string `query:"optional"`
}

type jsonDescriptor struct {
	Type string
	Data any
}

func (mod *Module) DescribeObject(ctx context.Context, objectID object.ID, scope *astral.Scope) (list []*objects.SourcedObject) {
	for _, d := range mod.describers.Clone() {
		list = append(list, d.DescribeObject(ctx, objectID, scope)...)
	}

	return
}

func (mod *Module) AddDescriber(describer objects.Describer) error {
	return mod.describers.Add(describer)
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

	return query.Accept(q, w, func(conn astral.Conn) {
		defer conn.Close()

		list := p.mod.DescribeObject(ctx, args.ID, astral.DefaultScope())
		list = slices.DeleteFunc(list, func(so *objects.SourcedObject) bool {
			return !p.mod.Auth.Authorize(q.Caller, objects.ActionAccessDescriptor, so.Object)
		})

		switch args.Format {
		case "":
			for _, so := range list {
				err = objects.WriteObject(conn, so.Object)
				if err != nil {
					p.mod.log.Errorv(1, "describe: error writing object: %v", err)
					return
				}
			}

		case "json":
			var d []jsonDescriptor
			for _, so := range list {
				d = append(d, jsonDescriptor{
					Type: so.Object.ObjectType(),
					Data: so.Object,
				})
			}
			err = json.NewEncoder(conn).Encode(d)
			if err != nil {
				p.mod.log.Errorv(1, "describe.json: error encoding json: %v", err)
				return
			}
		}
	})
}

func (c *Consumer) Describe(ctx context.Context, objectID object.ID, _ *astral.Scope) (list []*objects.SourcedObject, err error) {
	var q = query.New(
		c.consumerID,
		c.providerID,
		methodDescribe,
		&describeArgs{ID: objectID})

	conn, err := query.Route(ctx, c.mod.node, q)
	if err != nil {
		return
	}
	defer conn.Close()

	for {
		obj, err := c.mod.ReadObject(conn)
		if err != nil {
			if errors.Is(err, io.EOF) {
				break
			}
			return list, err
		}

		list = append(list, &objects.SourcedObject{
			Source: c.providerID,
			Object: obj,
		})
	}

	return
}

func (a *describeArgs) Validate() error {
	if a.ID.IsZero() {
		return errors.New("object ID is required")
	}
	switch a.Format {
	case "", "json":
	default:
		return errors.New("invlid format: " + a.Format)
	}
	return nil
}
