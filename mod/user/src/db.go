package user

type dbIdentity struct {
	Identity string `gorm:"primaryKey"`
}

func (dbIdentity) TableName() string {
	return "identities"
}
