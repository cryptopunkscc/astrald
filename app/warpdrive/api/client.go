package api

type Client interface {
	// Peers available for receiving an offer.
	Peers() ([]Peer, error)
	// Update peer attribute [alias, mod].
	Update(id PeerId, attr string, value string) error
	// Send files offer for the recipient.
	Send(peerId PeerId, path string) (OfferId, bool, error)
	// Download files from offer with given id.
	Download(id OfferId) error
	// Subscribe for new offers.
	Subscribe(filter Filter) (<-chan Offer, error)
	// Status channel for receiving offer updates
	Status(filter Filter) (<-chan Status, error)
	// Offers list
	Offers(filter Filter) ([]Offer, error)
}

type Filter string

const (
	FilterAll = Filter("all")
	FilterIn  = Filter("in")
	FilterOut = Filter("out")
)
