package objects

import (
	"errors"
	"sync"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/objects"
)

func (mod *Module) Describe(ctx *astral.Context, objectID *astral.ObjectID) (<-chan *objects.DescribeResult, error) {
	var results = make(chan *objects.DescribeResult)

	go func() {
		defer close(results)

		var wg sync.WaitGroup

		for _, d := range mod.describers.Clone() {
			d := d
			wg.Add(1)
			go func() {
				defer wg.Done()
				_res, _err := d.DescribeObject(ctx, objectID)
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
	ID   *astral.ObjectID
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
