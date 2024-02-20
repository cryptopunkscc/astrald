package media

import (
	"context"
	"github.com/cryptopunkscc/astrald/data"
	"github.com/cryptopunkscc/astrald/lib/desc"
	"github.com/cryptopunkscc/astrald/mod/media"
)

func (mod *Module) Describe(ctx context.Context, dataID data.ID, opts *desc.Opts) []*desc.Desc {
	var row dbMediaInfo
	var err = mod.db.Where("data_id = ?", dataID).First(&row).Error
	if err != nil {
		return nil
	}

	return []*desc.Desc{{
		Source: mod.node.Identity(),
		Data: media.Desc{
			MediaType: row.Type,
			Title:     row.Title,
			Artist:    row.Artist,
			Album:     row.Album,
			Genre:     row.Genre,
			Duration:  row.Duration,
		},
	}}
}
