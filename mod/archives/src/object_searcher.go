package archives

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"strings"
)

func (mod *Module) SearchObject(ctx *astral.Context, query string, opts *objects.SearchOpts) (<-chan *objects.SearchResult, error) {
	if !ctx.Zone().Is(astral.ZoneVirtual) {
		return nil, astral.ErrZoneExcluded
	}

	var results = make(chan *objects.SearchResult)

	go func() {
		defer close(results)

		var rows []*dbEntry

		query = "%" + strings.ToLower(query) + "%"

		err := mod.db.Where("LOWER(path) LIKE ?", query).Find(&rows).Error
		if err != nil {
			mod.log.Error("search: db: %v", err)
			return
		}

		for _, row := range rows {

			results <- &objects.SearchResult{
				ObjectID: row.ObjectID,
			}
		}
	}()

	return results, nil
}
