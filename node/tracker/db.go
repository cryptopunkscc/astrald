package tracker

import (
	"fmt"
	"github.com/cryptopunkscc/astrald/auth/id"
	"time"
)

type dbEndpoint struct {
	Identity  string `gorm:"primaryKey"`
	Network   string `gorm:"primaryKey"`
	Address   string `gorm:"primaryKey"`
	CreatedAt time.Time
	ExpiresAt time.Time
}

func (dbEndpoint) TableName() string {
	return "endpoints"
}

func (tracker *CoreTracker) dbEndpointToEndopoint(src dbEndpoint) (TrackedEndpoint, error) {
	identity, err := id.ParsePublicKeyHex(src.Identity)
	if err != nil {
		return TrackedEndpoint{}, err
	}

	endpoint, err := tracker.parser.Parse(src.Network, src.Address)
	if err != nil {
		fmt.Println(src.Network, src.Address)
		return TrackedEndpoint{}, err
	}

	return TrackedEndpoint{
		Identity:  identity,
		Endpoint:  endpoint,
		ExpiresAt: src.ExpiresAt,
	}, nil
}
