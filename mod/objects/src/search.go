package objects

import (
	"context"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"sync"
)

func (mod *Module) Search(ctx context.Context, query string, opts *objects.SearchOpts) (<-chan *objects.SearchResult, error) {
	if opts == nil {
		opts = objects.DefaultSearchOpts()
	}

	var results = make(chan *objects.SearchResult)

	go func() {
		defer close(results)

		var wg sync.WaitGroup

		for _, searcher := range mod.searchers.Clone() {
			searcher := searcher
			wg.Add(1)
			go func() {
				defer wg.Done()
				_res, _err := searcher.Search(ctx, query, opts)
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

func (mod *Module) AddSearcher(searcher objects.Searcher) error {
	return mod.searchers.Add(searcher)
}
