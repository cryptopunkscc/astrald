package content

import (
	"github.com/cryptopunkscc/astrald/mod/content"
	"github.com/cryptopunkscc/astrald/object"
)

type dbLabel struct {
	DataID object.ID `gorm:"primaryKey"`
	Label  string
}

func (dbLabel) TableName() string { return content.DBPrefix + "labels" }

func (mod *Module) SetLabel(objectID object.ID, label string) {
	mod.db.Where("data_id = ?", objectID).Delete(&dbLabel{})

	if label != "" {
		mod.db.Create(&dbLabel{
			DataID: objectID,
			Label:  label,
		})
	}
}

func (mod *Module) GetLabel(objectID object.ID) string {
	var label dbLabel

	tx := mod.db.Where("data_id = ?", objectID).First(&label)

	if tx.Error != nil {
		return ""
	}

	return label.Label
}
