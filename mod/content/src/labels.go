package content

import "github.com/cryptopunkscc/astrald/data"

type dbLabel struct {
	DataID string `gorm:"primaryKey,index"`
	Label  string
}

func (dbLabel) TableName() string { return "labels" }

func (mod *Module) SetLabel(dataID data.ID, label string) {
	mod.db.Where("data_id = ?", dataID.String()).Delete(&dbLabel{})

	if label != "" {
		mod.db.Create(&dbLabel{
			DataID: dataID.String(),
			Label:  label,
		})
	}
}

func (mod *Module) GetLabel(id data.ID) string {
	var label dbLabel

	tx := mod.db.Where("data_id = ?", id.String()).First(&label)

	if tx.Error != nil {
		return ""
	}

	return label.Label
}
