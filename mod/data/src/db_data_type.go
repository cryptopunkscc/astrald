package data

func (mod *Module) dbDataTypeFindByDataID(dataID string) (*dbDataType, error) {
	var row dbDataType
	var tx = mod.db.Where("data_id = ?", dataID).First(&row)
	return &row, tx.Error
}
