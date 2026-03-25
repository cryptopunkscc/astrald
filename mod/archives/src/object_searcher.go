package archives

import (
	"strings"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/objects"
)

func (mod *Module) SearchObject(ctx *astral.Context, query objects.SearchQuery) (<-chan *objects.SearchResult, error) {
	if !ctx.Zone().Is(astral.ZoneVirtual) {
		return nil, astral.ErrZoneExcluded
	}

	if !query.RequiredTagsIn("path", "archive") {
		return nil, objects.ErrTagNotSupported
	}

	var results = make(chan *objects.SearchResult)

	go func() {
		defer close(results)

		var rows []*dbEntry

		q := "%" + strings.ToLower(string(query.Query)) + "%"

		err := mod.db.Where("LOWER(path) LIKE ?", q).Find(&rows).Error
		if err != nil {
			mod.log.Error("search: db: %v", err)
			return
		}

		for _, row := range rows {
			results <- &objects.SearchResult{
				SourceID: mod.node.Identity(),
				ObjectID: row.ObjectID,
			}
		}
	}()

	return results, nil
}
