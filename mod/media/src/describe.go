package media

import (
	"context"
	"github.com/cryptopunkscc/astrald/data"
	"github.com/cryptopunkscc/astrald/mod/content"
	"github.com/cryptopunkscc/astrald/mod/media"
)

func (mod *Module) Describe(ctx context.Context, dataID data.ID, opts *content.DescribeOpts) []content.Descriptor {
	var row dbMediaInfo
	var err = mod.db.Where("data_id = ?", dataID).First(&row).Error
	if err != nil {
		return nil
	}

	return []content.Descriptor{
		media.Descriptor{
			Type:     row.Type,
			Title:    row.Title,
			Artist:   row.Artist,
			Album:    row.Album,
			Genre:    row.Genre,
			Duration: row.Duration,
		},
	}
}
