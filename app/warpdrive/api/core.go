package api

import (
	"io"
	"log"
)

// Core internal API for cache, storage, and subscription management.
type Core interface {
	Setup()
	SetLogger(logger *log.Logger)

	Peer() PeerManager
	Outgoing() OfferManager
	Incoming() OfferManager
	File() FileManager
}

type PeerManager interface {
	Update(peerId string, attr string, value string)
	Get(id PeerId) Peer
	List() (peers []*Peer)
	Offers() *Subscriptions
}

type OfferManager interface {
	Add(offerId string, files []Info, peer *Peer)
	Update(offer *Offer, status string, persist bool)
	Get(id OfferId) *Offer
	List() Offers
	Status() *Subscriptions
}

type FileManager interface {
	Info(uri string) (files []Info, err error)
	CopyFrom(reader io.Reader, offer *Offer) (err error)
	CopyTo(writer io.Writer, offer *Offer) (err error)
}
