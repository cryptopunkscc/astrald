package storage

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/data"
	"time"
)

type AccessManager interface {
	AccessVerifier
	Grant(identity id.Identity, dataID data.ID, expiresAt time.Time) error
	Revoke(identity id.Identity, dataID data.ID) error
	AddAccessVerifier(checker AccessVerifier)
	RemoveAccessVerifier(checker AccessVerifier)
}

type AccessVerifier interface {
	Verify(identity id.Identity, dataID data.ID) bool
}
