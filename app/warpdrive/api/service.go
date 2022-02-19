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
	Add(offerId string, files []Info, peerId PeerId)
	Update(offer *Offer, index int)
	Get(id OfferId) *Offer
	List() []Offer
	Status() *Subscriptions
}

type FileService interface {
	Info(uri string) (files []Info, err error)
	CopyFrom(reader io.Reader, offer *Offer) (err error)
	CopyTo(writer io.Writer, offer *Offer) (err error)
}
