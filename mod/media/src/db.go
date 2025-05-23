package media

import (
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/media"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"
)

type DB struct {
	*gorm.DB
}

func (db *DB) FindObject(objectID *astral.ObjectID) (row *dbObject, err error) {
	err = db.Where("object_id = ?", objectID).First(&row).Error
	return
}

func (db *DB) SaveObject(objectID *astral.ObjectID) error {
	return db.
		Clauses(clause.OnConflict{DoNothing: true}).
		Create(&dbObject{
			ObjectID: objectID,
		}).Error
}

func (db *DB) DeleteObject(objectID *astral.ObjectID) (err error) {
	return db.
		Delete(&dbObject{ObjectID: objectID}).
		Error
}

func (db *DB) FindAudio(objectID *astral.ObjectID) (row *dbAudio, err error) {
	err = db.Where("object_id = ?", objectID).First(&row).Error
	return
}

func (db *DB) SaveAudio(audio *media.AudioFile) error {
	return db.
		Clauses(clause.OnConflict{UpdateAll: true}).
		Create(&dbAudio{
			ObjectID:  audio.ObjectID,
			Format:    string(audio.Format),
			Title:     string(audio.Title),
			Artist:    string(audio.Artist),
			Album:     string(audio.Album),
			Genre:     string(audio.Genre),
			Year:      int(audio.Year),
			PictureID: audio.PictureID,
		}).Error
}

func (db *DB) DeleteAudio(objectID *astral.ObjectID) (err error) {
	return db.
		Delete(&dbAudio{ObjectID: objectID}).
		Error
}

func (db *DB) FindAudioContainerID(objectID *astral.ObjectID) (containerID *astral.ObjectID, err error) {
	err = db.
		Model(&dbAudio{}).
		Where("picture_id = ?", objectID).
		Select("object_id").
		First(&containerID).Error

	return
}

func (db *DB) UniquePictureIDs() (ids []*astral.ObjectID, err error) {
	err = db.
		Model(&dbAudio{}).
		Distinct("picture_id").
		Find(&ids).
		Error

	return
}
