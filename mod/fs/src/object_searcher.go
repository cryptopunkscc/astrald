package fs

import (
	"context"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"strings"
)

type Finder struct {
	mod *Module
}

func NewFinder(module *Module) *Finder {
	return &Finder{mod: module}
}

func (finder *Finder) Search(ctx context.Context, query string, opts *objects.SearchOpts) (<-chan *objects.SearchResult, error) {
	if !opts.Zone.Is(astral.ZoneDevice) {
		return nil, astral.ErrZoneExcluded
	}

	var results = make(chan *objects.SearchResult)

	go func() {
		defer close(results)

		var rows []*dbLocalFile

		err := finder.mod.db.
			Where("LOWER(PATH) like ?", "%"+strings.ToLower(query)+"%").
			Find(&rows).
			Error
		if err != nil {
			finder.mod.log.Error("search: db: %v", err)
			return
		}

		for _, row := range rows {
			results <- &objects.SearchResult{
				ObjectID: row.DataID,
			}
		}
	}()

	return results, nil
}
