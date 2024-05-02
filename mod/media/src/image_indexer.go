package media

import (
	"context"
	"github.com/cryptopunkscc/astrald/lib/desc"
	"github.com/cryptopunkscc/astrald/mod/media"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/object"
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

func (mod *ImageIndexer) Describe(ctx context.Context, objectID object.ID, opts *desc.Opts) []*desc.Desc {
	var row dbImage
	var err = mod.db.Where("data_id = ?", objectID).First(&row).Error
	var img *media.Image
	if err == nil {
		img = &media.Image{
			Format: row.Format,
			Width:  row.Width,
			Height: row.Height,
		}
	} else {
		img, err = mod.index(objectID, &objects.OpenOpts{
			Virtual: true,
		})
		if err != nil {
			mod.log.Errorv(2, "error indexing %v: %v", objectID, err)
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

func (mod *ImageIndexer) Find(ctx context.Context, query string, opts *objects.FindOpts) (matches []objects.Match, err error) {
	// images are not searchable yet
	return
}

func (mod *ImageIndexer) index(objectID object.ID, opts *objects.OpenOpts) (*media.Image, error) {
	r, err := mod.objects.Open(objectID, opts)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	img, err := mod.scan(r)
	if err != nil {
		return nil, err
	}

	err = mod.db.Create(&dbImage{
		DataID: objectID,
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
