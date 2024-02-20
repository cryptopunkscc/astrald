package media

import (
	"context"
	"github.com/cryptopunkscc/astrald/mod/content"
	"strings"
)

func (mod *Module) Find(ctx context.Context, query string, opts *content.FindOpts) ([]content.Match, error) {
	var rows []dbMediaInfo
	var matches []content.Match
	var pattern = "%" + strings.ToLower(query) + "%"

	var err = mod.db.
		Where(
			"LOWER(artist) like ? or LOWER(title) like ?",
			pattern,
			pattern,
		).
		Find(&rows).Error

	for _, row := range rows {
		matches = append(matches, content.Match{
			DataID: row.DataID,
			Score:  100,
			Exp:    "artist/title matches",
		})
	}

	return matches, err
}
