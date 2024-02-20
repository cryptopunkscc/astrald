package content

import (
	"context"
	"github.com/cryptopunkscc/astrald/mod/content"
)

func (mod *Module) Find(ctx context.Context, query string, opts *content.FindOpts) ([]content.Match, error) {
	var matches []content.Match
	var errs []error

	if opts == nil {
		opts = &content.FindOpts{}
	}

	for _, finder := range mod.finders.Clone() {
		m, err := finder.Find(ctx, query, opts)
		if err != nil {
			errs = append(errs, err)
		}
		matches = append(matches, m...)
	}

	return matches, nil
}

func (mod *Module) AddFinder(finder content.Finder) error {
	return mod.finders.Add(finder)
}

func (mod *Module) RemoveFinder(finder content.Finder) error {
	return mod.finders.Remove(finder)
}
