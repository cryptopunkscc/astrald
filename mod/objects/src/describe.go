package objects

import (
	"errors"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/object"
	"sync"
)

func (mod *Module) Describe(ctx *astral.Context, objectID *object.ID, scope *astral.Scope) (<-chan *objects.SourcedObject, error) {
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

type describeArgs struct {
	ID   *object.ID
	Out  string `query:"optional"`
	Zone string `query:"optional"`
}

func (a *describeArgs) Validate() error {
	if a.ID.IsZero() {
		return errors.New("object ID is required")
	}
	switch a.Out {
	case "", "json":
	default:
		return errors.New("invlid format: " + a.Out)
	}
	return nil
}
