package relay

import "time"

type dbRelayCert struct {
	DataID    string    `gorm:"primaryKey"`
	TargetID  string    `gorm:"index"`
	RelayID   string    `gorm:"index"`
	Direction string    `gorm:"index"`
	ExpiresAt time.Time `gorm:"index"`
}

func (dbRelayCert) TableName() string {
	return "relay_certs"
}
