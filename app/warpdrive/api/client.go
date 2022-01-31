package api

// Sender API for warpdrive UI.
type Sender interface {
	StatusEvents
	// Peers available for receiving an offer.
	Peers() ([]Peer, error)
	// Send files offer for the recipient.
	Send(peerId PeerId, path string) (OfferId, ResponseCode, error)
	// Sent offers.
	Sent() (Offers, error)
	Status(id OfferId) (string, error)
}

// Recipient API for warpdrive UI.
type Recipient interface {
	StatusEvents
	// Offers subscription.
	Offers() (<-chan Offer, error)
	// Received offers.
	Received() (Offers, error)
	// Accept offer and starts in background downloading.
	Accept(id OfferId) error
	// Reject offer.
	Reject(id OfferId) error
	// Update peer attribute [alias, mod].
	Update(id PeerId, attr string, value string) error
}

type StatusEvents interface {
	// Events subscribes a callback for receiving offers status updates.
	Events() (<-chan Status, error)
}
