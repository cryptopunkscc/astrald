package uid

import "github.com/cryptopunkscc/astrald/api"

type Identities interface {
	Update(card Card) error
	List() ([]Card, error)
	Get(identity api.Identity) (*Card, error)
}

type Card struct {
	Id        api.Identity
	Alias     string
	Endpoints []Endpoint
}

type Endpoint struct {
	Network string
	Address string
}
