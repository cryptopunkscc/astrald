package media

import (
	"context"
	"github.com/cryptopunkscc/astrald/lib/desc"
	"github.com/cryptopunkscc/astrald/mod/media"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/object"
	"github.com/dhowden/tag"
	"strings"
)

var _ Indexer = &AudioIndexer{}

type AudioIndexer struct {
	*Module
}

func NewAudioIndexer(mod *Module) *AudioIndexer {
	return &AudioIndexer{Module: mod}
}

func (mod *AudioIndexer) Describe(ctx context.Context, objectID object.ID, opts *desc.Opts) (descs []*desc.Desc) {
	openOpts := &objects.OpenOpts{
		Zone: astral.ZoneDevice | astral.ZoneVirtual,
	}

	if opts.Zone.Is(astral.ZoneNetwork) {
		openOpts.Zone |= astral.ZoneNetwork
	}

	audio, _ := mod.Index(ctx, objectID, openOpts)

	if audio != nil {
		descs = append(descs, &desc.Desc{
			Source: mod.node.Identity(),
			Data:   audio,
		})
	}

	return
}

func (mod *AudioIndexer) Search(ctx context.Context, query string, opts *objects.SearchOpts) (matches []objects.Match, err error) {
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
			ObjectID: row.ObjectID,
			Score:    100,
			Exp:      "audio tags match query",
		})
	}

	return
}

func (mod *AudioIndexer) Forget(objectID object.ID) error {
	return mod.clearCache(objectID)
}

func (mod *AudioIndexer) Index(ctx context.Context, objectID object.ID, opts *objects.OpenOpts) (*media.Audio, error) {
	// check cache
	if c := mod.getCache(objectID); c != nil {
		return c, nil
	}

	// scan the object
	info, err := mod.scanObject(ctx, objectID, opts)
	if err != nil {
		return nil, err
	}

	// save to cache
	return info, mod.setCache(objectID, info)
}

func (mod *AudioIndexer) scanObject(ctx context.Context, objectID object.ID, opts *objects.OpenOpts) (*media.Audio, error) {
	r, err := mod.objects.Open(ctx, objectID, opts)
	if err != nil {
		return nil, err
	}
	defer r.Close()

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

func (mod *AudioIndexer) setCache(objectID object.ID, audio *media.Audio) error {
	return mod.db.Create(&dbAudio{
		ObjectID: objectID,
		Format:   audio.Format,
		Duration: audio.Duration,
		Title:    audio.Title,
		Artist:   audio.Artist,
		Album:    audio.Album,
		Genre:    audio.Genre,
		Year:     audio.Year,
	}).Error
}

func (mod *AudioIndexer) clearCache(objectID object.ID) error {
	return mod.db.
		Where("object_id = ?", objectID).
		Delete(&dbAudio{}).
		Error
}

func (mod *AudioIndexer) getCache(objectID object.ID) (audio *media.Audio) {
	var row dbAudio

	err := mod.db.Where("object_id = ?", objectID).First(&row).Error
	if err != nil {
		return nil
	}

	return &media.Audio{
		Format:   row.Format,
		Duration: row.Duration,
		Title:    row.Title,
		Artist:   row.Artist,
		Album:    row.Album,
		Genre:    row.Genre,
		Year:     row.Year,
	}
}
