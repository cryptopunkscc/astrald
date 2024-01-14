package storage

import (
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/data"
)

type AccessManager interface {
	AccessVerifier
	AddAccessVerifier(checker AccessVerifier)
	RemoveAccessVerifier(checker AccessVerifier)
}

type AccessVerifier interface {
	Verify(identity id.Identity, dataID data.ID) bool
}
