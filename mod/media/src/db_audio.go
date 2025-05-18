package media

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/media"
	"time"
)

type dbAudio struct {
	ObjectID  *astral.ObjectID `gorm:"primaryKey"`
	Format    string           `gorm:"index"`
	Duration  time.Duration
	Title     string           `gorm:"index"`
	Artist    string           `gorm:"index"`
	Album     string           `gorm:"index"`
	Genre     string           `gorm:"index"`
	Year      int              `gorm:"index"`
	PictureID *astral.ObjectID `gorm:"index"`
}

func (dbAudio) TableName() string { return media.DBPrefix + "audio" }

func (row *dbAudio) ToAudioFile() *media.AudioFile {
	return &media.AudioFile{
		ObjectID:  row.ObjectID,
		Format:    astral.String8(row.Format),
		Title:     astral.String8(row.Title),
		Artist:    astral.String8(row.Artist),
		Album:     astral.String8(row.Album),
		Genre:     astral.String8(row.Genre),
		Year:      astral.Uint16(row.Year),
		PictureID: row.PictureID,
	}
}
