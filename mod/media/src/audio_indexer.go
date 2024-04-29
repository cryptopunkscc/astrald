package media

import (
	"context"
	"github.com/cryptopunkscc/astrald/lib/desc"
	"github.com/cryptopunkscc/astrald/mod/media"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/object"
	"github.com/dhowden/tag"
	"io"
	"strings"
)

var _ Indexer = &AudioIndexer{}

type AudioIndexer struct {
	*Module
}

func NewAudioIndexer(mod *Module) *AudioIndexer {
	return &AudioIndexer{Module: mod}
}

func (mod *AudioIndexer) Describe(ctx context.Context, objectID object.ID, opts *desc.Opts) []*desc.Desc {
	var audio *media.Audio
	var row dbAudio
	var err = mod.db.Where("data_id = ?", objectID).First(&row).Error
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
		audio, err = mod.index(objectID, &objects.OpenOpts{
			Virtual: true,
		})
		if err != nil {
			mod.log.Errorv(2, "error indexing %v: %v", objectID, err)
		} else {
			mod.log.Infov(1, "indexed %s by %s", audio.Title, audio.Artist)
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

func (mod *AudioIndexer) Find(ctx context.Context, query string, opts *objects.FindOpts) (matches []objects.Match, err error) {
	var rows []*dbAudio

	query = "%" + strings.ToLower(query) + "%"

	err = mod.db.
		Where("LOWER(artist) LIKE ? OR LOWER(title) LIKE ? OR LOWER(album) LIKE ?", query, query, query).
		Find(&rows).
		Error
	if err != nil {
		mod.log.Error("db error: %v", err)
		return
	}

	for _, row := range rows {
		matches = append(matches, objects.Match{
			ObjectID: row.DataID,
			Score:    100,
			Exp:      "audio tags match query",
		})
	}

	return
}

func (mod *AudioIndexer) index(objectID object.ID, opts *objects.OpenOpts) (*media.Audio, error) {
	r, err := mod.objects.Open(objectID, opts)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	info, err := mod.scan(r)
	if err != nil {
		return nil, err
	}

	err = mod.db.Create(&dbAudio{
		DataID:   objectID,
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
