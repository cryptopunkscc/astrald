package fs

import (
	"context"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"strings"
)

type Finder struct {
	mod *Module
}

func NewFinder(module *Module) *Finder {
	return &Finder{mod: module}
}

func (finder *Finder) Find(ctx context.Context, query string, opts *objects.FindOpts) (matches []objects.Match, err error) {
	var rows []*dbLocalFile

	err = finder.mod.db.
		Where("LOWER(PATH) like ?", "%"+strings.ToLower(query)+"%").
		Find(&rows).
		Error
	if err != nil {
		return
	}

	for _, row := range rows {
		matches = append(matches, objects.Match{
			ObjectID: row.DataID,
			Score:    100,
			Exp:      "file path contains query",
		})
	}

	return
}
