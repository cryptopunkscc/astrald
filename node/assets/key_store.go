package assets

import "github.com/cryptopunkscc/astrald/auth/id"

type KeyStore interface {
	Save(identity id.Identity) error
	Find(identity id.Identity) (id.Identity, error)
}
