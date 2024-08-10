package objects

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/lib/query"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/object"
	"io"
)

type describeArgs struct {
	ObjectID object.ID
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

	if !p.mod.Auth.Authorize(q.Caller, objects.ActionRead, &args.ObjectID) {
		return query.Reject()
	}

	return query.Accept(q, w, func(conn astral.Conn) {
		defer conn.Close()

		for _, so := range p.mod.DescribeObject(ctx, args.ObjectID, astral.DefaultScope()) {
			if !p.mod.Auth.Authorize(q.Caller, objects.ActionAccessDescriptor, so.Object) {
				continue
			}

			err = objects.WriteObject(conn, so.Object)
			if err != nil {
				p.mod.log.Errorv(1, "describe: error writing object: %v", err)
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
		&describeArgs{ObjectID: objectID})

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
