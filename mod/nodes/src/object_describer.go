package nodes

import (
	"slices"
	"sync"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/objects"
	objectscli "github.com/cryptopunkscc/astrald/mod/objects/client"
)

func (mod *Module) DescribeObject(ctx *astral.Context, objectID *astral.ObjectID) (<-chan *objects.DescribeResult, error) {
	if !ctx.Zone().Is(astral.ZoneNetwork) {
		return nil, astral.ErrZoneExcluded
	}

	var results = make(chan *objects.DescribeResult)

	go func() {
		defer close(results)

		providers := mod.FindObject(ctx, objectID)

		providers = slices.DeleteFunc(providers, func(identity *astral.Identity) bool {
			return mod.Dir.ApplyFilters(identity, ctx.Filters()...)
		})

		var wg sync.WaitGroup

		for _, providerID := range providers {
			providerID := providerID

			wg.Add(1)
			go func() {
				defer wg.Done()

				_results, err := objectscli.New(providerID, nil).Describe(ctx, objectID)
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
