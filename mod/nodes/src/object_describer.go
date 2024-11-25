package nodes

import (
	"context"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/object"
	"slices"
	"sync"
)

func (mod *Module) DescribeObject(ctx context.Context, objectID object.ID, scope *astral.Scope) (<-chan *objects.SourcedObject, error) {
	if !scope.Zone.Is(astral.ZoneNetwork) {
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

				c, err := mod.Objects.Connect(providerID, nil)
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
