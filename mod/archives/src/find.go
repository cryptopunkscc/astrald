package archives

import (
	"context"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"strings"
)

func (mod *Module) Find(ctx context.Context, query string, opts *objects.FindOpts) (matches []objects.Match, err error) {
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
