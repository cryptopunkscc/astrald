package contacts

import "github.com/cryptopunkscc/astrald/auth/id"

type Contacts interface {
	DisplayName(nodeID id.Identity) string
	Find(nodeID id.Identity) (c *Contact, err error)
	FindOrCreate(nodeID id.Identity) (c *Contact, err error)
	FindByAlias(alias string) (c *Contact, found bool)
	Delete(identity id.Identity) error
	ResolveIdentity(str string) (id.Identity, error)
	All() <-chan *Contact
}
