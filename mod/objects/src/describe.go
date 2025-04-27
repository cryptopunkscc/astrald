package objects

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/lib/query"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/object"
	"sync"
)

func (mod *Module) Describe(ctx *astral.Context, objectID object.ID, scope *astral.Scope) (<-chan *objects.SourcedObject, error) {
	if scope == nil {
		scope = astral.DefaultScope()
	}

	var results = make(chan *objects.SourcedObject)

	go func() {
		defer close(results)

		var wg sync.WaitGroup

		for _, d := range mod.describers.Clone() {
			d := d
			wg.Add(1)
			go func() {
				defer wg.Done()
				_res, _err := d.DescribeObject(ctx, objectID, scope)
				if _err != nil {
					return
				}

				for i := range _res {
					results <- i
				}
			}()
		}

		wg.Wait()
	}()

	return results, nil
}

func (mod *Module) AddDescriber(describer objects.Describer) error {
	return mod.describers.Add(describer)
}

func (c *Consumer) Describe(ctx context.Context, objectID object.ID, _ *astral.Scope) (<-chan *objects.SourcedObject, error) {
	var results = make(chan *objects.SourcedObject, 1)

	var q = query.New(
		c.consumerID,
		c.providerID,
		methodDescribe,
		&describeArgs{ID: objectID})

	go func() {
		defer close(results)

		conn, err := query.Route(ctx, c.mod.node, q)
		if err != nil {
			return
		}
		defer conn.Close()

		for {
			obj, _, err := c.mod.Blueprints().Read(conn, true)
			if err != nil {
				return
			}

			results <- &objects.SourcedObject{
				Source: c.providerID,
				Object: obj,
			}
		}
	}()

	return results, nil
}

type describeArgs struct {
	ID     object.ID
	Format string `query:"optional"`
	Zones  string `query:"optional"`
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
