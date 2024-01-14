package keys

type dbPrivateKey struct {
	DataID    string `gorm:"uniqueIndex"`
	Type      string `gorm:"index"`
	PublicKey string `gorm:"index"`
}

func (dbPrivateKey) TableName() string {
	return "private_keys"
}
