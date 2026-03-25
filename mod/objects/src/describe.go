package objects

import (
	"sync"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/objects"
)

func (mod *Module) Describe(ctx *astral.Context, objectID *astral.ObjectID) (<-chan *objects.Descriptor, error) {
	var results = make(chan *objects.Descriptor)

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
