package nodes

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"slices"
	"sync"
)

func (mod *Module) DescribeObject(ctx *astral.Context, objectID *astral.ObjectID, scope *astral.Scope) (<-chan *objects.SourcedObject, error) {
	if !ctx.Zone().Is(astral.ZoneNetwork) {
		return nil, astral.ErrZoneExcluded
	}

	var results = make(chan *objects.SourcedObject)

	go func() {
		defer close(results)

		providers := mod.FindObject(ctx, objectID, scope)

		if scope.QueryFilter != nil {
			providers = slices.DeleteFunc(providers, func(identity *astral.Identity) bool {
				return !scope.QueryFilter(identity)
			})
		}

		var wg sync.WaitGroup

		for _, providerID := range providers {
			providerID := providerID

			wg.Add(1)
			go func() {
				defer wg.Done()

				c, err := mod.Objects.On(providerID, nil)
				if err != nil {
					return
				}

				_results, err := c.Describe(ctx, objectID, scope)
				if err != nil {
					return
				}

				for r := range _results {
					results <- r
				}
			}()
		}

		wg.Wait()
	}()

	return results, nil
}
