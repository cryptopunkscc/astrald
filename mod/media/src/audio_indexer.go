package media

import (
	"bytes"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/media"
	"github.com/cryptopunkscc/astrald/mod/objects"
	"github.com/dhowden/tag"
)

var _ Indexer = &AudioIndexer{}

type AudioIndexer struct {
	*Module
}

func NewAudioIndexer(mod *Module) *AudioIndexer {
	return &AudioIndexer{Module: mod}
}

func (mod *AudioIndexer) Index(ctx *astral.Context, objectID *astral.ObjectID) (f *media.AudioFile, err error) {
	// check if already indexed
	if row, err := mod.db.FindAudio(objectID); err == nil {
		return row.ToAudioFile(), nil
	}

	// inspect the object
	audio, err := mod.Inspect(ctx, objectID)
	if err != nil {
		return nil, err
	}

	if !audio.PictureID.IsZero() {
		mod.repo.push(audio.PictureID)
	}

	mod.log.Infov(1, "indexed %v by %v", audio.Title, audio.Artist)

	// save to cache
	return audio, mod.db.SaveAudio(audio)
}

func (mod *AudioIndexer) Inspect(ctx *astral.Context, objectID *astral.ObjectID) (*media.AudioFile, error) {
	// open the object
	r, err := mod.Objects.ReadDefault().Read(ctx, objectID, 0, 0)
	if err != nil {
		return nil, err
	}
	defer r.Close()

	// read id3 tag data
	audioTag, err := tag.ReadFrom(objects.NewReadSeeker(ctx, objectID, mod.Objects.ReadDefault(), r))
	if err != nil {
		return nil, err
	}

	// read the cover image
	var pictureID *astral.ObjectID
	if p := audioTag.Picture(); p != nil {
		pictureID, err = astral.Resolve(bytes.NewReader(p.Data))
	}

	// return info
	return &media.AudioFile{
		ObjectID:  objectID,
		Format:    astral.String8(audioTag.FileType()),
		Title:     astral.String8(audioTag.Title()),
		Artist:    astral.String8(audioTag.Artist()),
		Album:     astral.String8(audioTag.Album()),
		Genre:     astral.String8(audioTag.Genre()),
		Year:      astral.Uint16(audioTag.Year()),
		PictureID: pictureID,
	}, err
}
