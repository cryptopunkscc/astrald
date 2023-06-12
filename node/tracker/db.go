package tracker

import (
	"github.com/cryptopunkscc/astrald/net"
	"time"
)

type dbEndpoint struct {
	Identity      string `gorm:"primaryKey"`
	Network       string `gorm:"primaryKey"`
	Address       string `gorm:"primaryKey"`
	CreatedAt     time.Time
	LastSuccessAt time.Time
}

func (dbEndpoint) TableName() string { return "endpoints" }

func (tracker *CoreTracker) dbEndpointToEndopoint(src dbEndpoint) (net.Endpoint, error) {
	endpoint, err := tracker.parser.Parse(src.Network, src.Address)
	if err != nil {
		return nil, err
	}

	return endpoint, nil
}

func (tracker *CoreTracker) dbAutoMigrate() (err error) {
	return tracker.db.AutoMigrate(
		&dbEndpoint{},
		&dbAliases{},
	)
}

type dbAliases struct {
	Identity string `gorm:"primaryKey"`
	Alias    string `gorm:"index;unique;not null"`
}

func (dbAliases) TableName() string { return "aliases" }
