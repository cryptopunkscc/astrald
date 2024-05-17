package objects

import (
	"context"
	"github.com/cryptopunkscc/astrald/mod/objects"
)

func (mod *Module) Search(ctx context.Context, query string, opts *objects.SearchOpts) ([]objects.Match, error) {
	var matches []objects.Match
	var errs []error

	if opts == nil {
		opts = objects.DefaultSearchOpts()
	}

	for _, searcher := range mod.searchers.Clone() {
		m, err := searcher.Search(ctx, query, opts)
		if err != nil {
			errs = append(errs, err)
		}
		matches = append(matches, m...)
	}

	return matches, nil
}

func (mod *Module) AddSearcher(searcher objects.Searcher) error {
	return mod.searchers.Add(searcher)
}
