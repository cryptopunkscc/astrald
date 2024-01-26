package media

import (
	"context"
	_data "github.com/cryptopunkscc/astrald/data"
	"github.com/cryptopunkscc/astrald/mod/data"
	"github.com/cryptopunkscc/astrald/mod/media"
)

func (mod *Module) DescribeData(ctx context.Context, dataID _data.ID, opts *data.DescribeOpts) []data.Descriptor {
	var row dbMediaInfo
	var err = mod.db.Where("data_id = ?", dataID).First(&row).Error
	if err != nil {
		return nil
	}

	return []data.Descriptor{{
		Type: media.MediaDescriptorType,
		Data: media.MediaDescriptor{
			Type:     row.Type,
			Title:    row.Title,
			Artist:   row.Artist,
			Album:    row.Album,
			Genre:    row.Genre,
			Duration: row.Duration,
		},
	}}
}
