package api

import (
	"io"
)

type PeerService interface {
	Update(peerId string, attr string, value string)
	Get(id PeerId) Peer
	List() []Peer
	Offers() *Subscriptions
}

type OfferService interface {
	Add(offerId string, files []Info, peer *Peer)
	Update(offer *Offer, index int)
	Get() Offers
	Status() *Subscriptions
}

type FileService interface {
	Info(uri string) (files []Info, err error)
	CopyFrom(reader io.Reader, offer *Offer) (err error)
	CopyTo(writer io.Writer, offer *Offer) (err error)
}
