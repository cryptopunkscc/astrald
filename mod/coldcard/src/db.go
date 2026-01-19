package coldcard

import "gorm.io/gorm"

type DB struct {
	*gorm.DB
}

func newDB(gormDB *gorm.DB) (*DB, error) {
	db := &DB{DB: gormDB}

	err := db.DB.AutoMigrate()
	if err != nil {
		return nil, err
	}

	return db, nil
}
