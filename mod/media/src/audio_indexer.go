package media

import (
	"bytes"
	"context"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/media"
	"github.com/cryptopunkscc/astrald/mod/objects"
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

func (mod *AudioIndexer) DescribeObject(ctx context.Context, objectID object.ID, opts *astral.Scope) (list []*objects.SourcedObject) {
	openOpts := &objects.OpenOpts{
		Zone: astral.ZoneDevice | astral.ZoneVirtual,
	}

	if opts.Zone.Is(astral.ZoneNetwork) {
		openOpts.Zone |= astral.ZoneNetwork
	}

	audio, _ := mod.Index(ctx, objectID, openOpts)

	if audio == nil {
		return
	}

	list = append(list, &objects.SourcedObject{
		Source: mod.node.Identity(),
		Object: audio,
	})

	return
}

func (mod *AudioIndexer) Search(ctx context.Context, query string, opts *objects.SearchOpts) (<-chan *objects.SearchResult, error) {
	if !opts.Zone.Is(astral.ZoneDevice) {
		return nil, astral.ErrZoneExcluded
	}
	var results = make(chan *objects.SearchResult)

	go func() {
		defer close(results)

		var rows []*dbAudio

		query = "%" + strings.ToLower(query) + "%"

		err := mod.db.
			Where("LOWER(artist) LIKE ? OR LOWER(title) LIKE ? OR LOWER(album) LIKE ?", query, query, query).
			Find(&rows).
			Error
		if err != nil {
			mod.log.Error("search: db : %v", err)
			return
		}

		for _, row := range rows {
			results <- &objects.SearchResult{
				ObjectID: row.ObjectID,
			}
		}
	}()

	return results, nil
}

func (mod *AudioIndexer) Forget(objectID object.ID) error {
	return mod.clearCache(objectID)
}

func (mod *AudioIndexer) Index(ctx context.Context, objectID object.ID, opts *objects.OpenOpts) (*media.AudioDescriptor, error) {
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

func (mod *AudioIndexer) scanObject(ctx context.Context, objectID object.ID, opts *objects.OpenOpts) (*media.AudioDescriptor, error) {
	r, err := mod.Objects.Open(ctx, objectID, opts)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	meta, err := tag.ReadFrom(r)
	if err != nil {
		return nil, err
	}

	var pictureID object.ID

	if p := meta.Picture(); p != nil {
		pictureID, err = object.Resolve(bytes.NewReader(p.Data))
	}

	return &media.AudioDescriptor{
		Format:  string(meta.FileType()),
		Title:   meta.Title(),
		Artist:  meta.Artist(),
		Album:   meta.Album(),
		Genre:   meta.Genre(),
		Year:    meta.Year(),
		Picture: pictureID,
	}, err
}

func (mod *AudioIndexer) setCache(objectID object.ID, audio *media.AudioDescriptor) error {
	return mod.db.Create(&dbAudio{
		ObjectID:  objectID,
		Format:    audio.Format,
		Title:     audio.Title,
		Artist:    audio.Artist,
		Album:     audio.Album,
		Genre:     audio.Genre,
		Year:      audio.Year,
		PictureID: audio.Picture,
	}).Error
}

func (mod *AudioIndexer) clearCache(objectID object.ID) error {
	return mod.db.
		Where("object_id = ?", objectID).
		Delete(&dbAudio{}).
		Error
}

func (mod *AudioIndexer) getCache(objectID object.ID) (audio *media.AudioDescriptor) {
	var row dbAudio

	err := mod.db.Where("object_id = ?", objectID).First(&row).Error
	if err != nil {
		return nil
	}

	return &media.AudioDescriptor{
		Format:  row.Format,
		Title:   row.Title,
		Artist:  row.Artist,
		Album:   row.Album,
		Genre:   row.Genre,
		Year:    row.Year,
		Picture: row.PictureID,
	}
}
