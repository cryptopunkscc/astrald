package archives

import (
	"context"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/astral"
	"strings"
)

func (mod *Module) Search(ctx context.Context, query string, opts *objects.SearchOpts) (matches []objects.Match, err error) {
	if !opts.Zone.Is(astral.ZoneVirtual) {
		return nil, astral.ErrZoneExcluded
	}

	var rows []*dbEntry

	query = "%" + strings.ToLower(query) + "%"

	err = mod.db.Where("LOWER(path) LIKE ?", query).Find(&rows).Error

	for _, row := range rows {
		matches = append(matches, objects.Match{
			ObjectID: row.ObjectID,
			Score:    50,
			Exp:      "compressed file name matches the query",
		})
	}

	return
}
