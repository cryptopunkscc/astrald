package media

import (
	"context"
	"github.com/cryptopunkscc/astrald/data"
	"github.com/cryptopunkscc/astrald/lib/desc"
	"github.com/cryptopunkscc/astrald/mod/media"
	"github.com/cryptopunkscc/astrald/mod/storage"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
)

var _ Indexer = &ImageIndexer{}

type ImageIndexer struct {
	*Module
}

func NewImageIndexer(mod *Module) *ImageIndexer {
	return &ImageIndexer{Module: mod}
}

func (mod *ImageIndexer) Describe(ctx context.Context, dataID data.ID, opts *desc.Opts) []*desc.Desc {
	var row dbImage
	var err = mod.db.Where("data_id = ?", dataID).First(&row).Error
	var img *media.Image
	if err == nil {
		img = &media.Image{
			Format: row.Format,
			Width:  row.Width,
			Height: row.Height,
		}
	} else {
		img, err = mod.index(dataID, &storage.OpenOpts{
			Virtual: true,
		})
		if err != nil {
			mod.log.Errorv(2, "error indexing %v: %v", dataID, err)
		}
	}
	if img == nil {
		return nil
	}

	return []*desc.Desc{{
		Source: mod.node.Identity(),
		Data:   img,
	}}
}

func (mod *ImageIndexer) index(dataID data.ID, opts *storage.OpenOpts) (*media.Image, error) {
	r, err := mod.storage.Open(dataID, opts)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	img, err := mod.scan(r)
	if err != nil {
		return nil, err
	}

	err = mod.db.Create(&dbImage{
		DataID: dataID,
		Format: img.Format,
		Width:  img.Width,
		Height: img.Height,
	}).Error

	return img, err
}

func (mod *ImageIndexer) scan(r io.Reader) (*media.Image, error) {
	i, f, err := image.DecodeConfig(r)
	if err != nil {
		return nil, err
	}

	return &media.Image{
		Format: f,
		Width:  i.Width,
		Height: i.Height,
	}, nil
}
