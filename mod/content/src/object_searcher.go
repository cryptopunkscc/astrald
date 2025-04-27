package content

import (
	"context"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"strings"
)

func (mod *Module) SearchObject(_ context.Context, query string, opts *objects.SearchOpts) (<-chan *objects.SearchResult, error) {
	if !opts.Zone.Is(astral.ZoneDevice) {
		return nil, astral.ErrZoneExcluded
	}

	var results = make(chan *objects.SearchResult)

	go func() {
		defer close(results)

		var rows []*dbDataType

		err := mod.db.
			Where("LOWER(TYPE) like ?", "%"+strings.ToLower(query)+"%").
			Find(&rows).
			Error
		if err != nil {
			mod.log.Error("search: db: %v", err)
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
