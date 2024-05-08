package archives

import (
	"context"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/net"
	"strings"
)

func (mod *Module) Find(ctx context.Context, query string, opts *objects.FindOpts) (matches []objects.Match, err error) {
	if !opts.Zone.Is(net.ZoneVirtual) {
		return nil, net.ErrZoneExcluded
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
