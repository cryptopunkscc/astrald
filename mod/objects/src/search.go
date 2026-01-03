package objects

import (
	"sync"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/lib/astrald"
	"github.com/cryptopunkscc/astrald/mod/objects"
)

func (mod *Module) Search(ctx *astral.Context, query string, opts *objects.SearchOpts) (<-chan *objects.SearchResult, error) {
	if opts == nil {
		opts = objects.DefaultSearchOpts()
	}

	search := &objects.Search{
		CallerID: ctx.Identity(),
		Query:    query,
	}

	for _, pre := range mod.searchPre.Clone() {
		pre.PreprocessSearch(search)
	}

	var results = make(chan *objects.SearchResult)
	var wg sync.WaitGroup

	// run local searchers
	for _, searcher := range mod.searchers.Clone() {
		searcher := searcher
		wg.Add(1)
		go func() {
			defer wg.Done()

			_res, _err := searcher.SearchObject(ctx, query, opts)
			if _err != nil {
				return
			}

			for r := range _res {
				select {
				case results <- r:
				case <-ctx.Done():
					return
				}
			}
		}()
	}

	// run network searchers
	if ctx.Zone().Is(astral.ZoneNetwork) {
		for _, nodeID := range search.Sources {
			nodeID := nodeID

			wg.Add(1)
			go func() {
				defer wg.Done()

				// execute search
				_results, err := astrald.NewObjectsClient(nodeID, nil).Search(ctx, query)
				if err != nil {
					mod.log.Errorv(1, "search %v: %v", nodeID, err)
					return
				}

				// copy results
				for r := range _results {
					select {
					case results <- r:
					case <-ctx.Done():
						return
					}
				}
			}()
		}
	}

	go func() {
		wg.Wait()
		close(results)
	}()

	return results, nil
}

func (mod *Module) AddSearcher(searcher objects.Searcher) error {
	return mod.searchers.Add(searcher)
}
