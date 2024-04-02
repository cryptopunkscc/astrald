package media

import (
	"context"
	"github.com/cryptopunkscc/astrald/data"
	"github.com/cryptopunkscc/astrald/lib/desc"
	"github.com/cryptopunkscc/astrald/mod/media"
	"github.com/cryptopunkscc/astrald/mod/storage"
	"github.com/dhowden/tag"
	"io"
)

var _ Indexer = &AudioIndexer{}

type AudioIndexer struct {
	*Module
}

func NewAudioIndexer(mod *Module) *AudioIndexer {
	return &AudioIndexer{Module: mod}
}

func (mod *AudioIndexer) Describe(ctx context.Context, dataID data.ID, opts *desc.Opts) []*desc.Desc {
	var audio *media.Audio
	var row dbAudio
	var err = mod.db.Where("data_id = ?", dataID).First(&row).Error
	if err == nil {
		audio = &media.Audio{
			Format:   row.Format,
			Duration: row.Duration,
			Title:    row.Title,
			Artist:   row.Artist,
			Album:    row.Album,
			Genre:    row.Genre,
			Year:     row.Year,
		}
	} else {
		audio, err = mod.index(dataID, &storage.OpenOpts{
			Virtual: true,
		})
		if err != nil {
			mod.log.Errorv(2, "error indexing %v: %v", dataID, err)
		}
	}
	if audio == nil {
		return nil
	}

	return []*desc.Desc{{
		Source: mod.node.Identity(),
		Data:   audio,
	}}
}

func (mod *AudioIndexer) index(dataID data.ID, opts *storage.OpenOpts) (*media.Audio, error) {
	r, err := mod.storage.Open(dataID, opts)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	info, err := mod.scan(r)
	if err != nil {
		return nil, err
	}

	err = mod.db.Create(&dbAudio{
		DataID:   dataID,
		Format:   info.Format,
		Duration: info.Duration,
		Title:    info.Title,
		Artist:   info.Artist,
		Album:    info.Album,
		Genre:    info.Genre,
		Year:     info.Year,
	}).Error

	return info, err
}

func (mod *AudioIndexer) scan(r io.ReadSeeker) (*media.Audio, error) {
	meta, err := tag.ReadFrom(r)
	if err != nil {
		return nil, err
	}

	return &media.Audio{
		Format: string(meta.FileType()),
		Title:  meta.Title(),
		Artist: meta.Artist(),
		Album:  meta.Album(),
		Genre:  meta.Genre(),
		Year:   meta.Year(),
	}, err
}
