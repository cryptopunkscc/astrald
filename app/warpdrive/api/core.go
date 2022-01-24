package api

import (
	"io"
	"log"
)

// Core internal API for cache, storage, and subscription management.
type Core interface {
	Resolver

	Setup()
	SetLogger(logger *log.Logger)

	UpdatePeer(peerId string, attr string, value string)
	GetPeer(id PeerId) Peer
	ListPeers() (peers []*Peer)

	AddOutgoingOffer(offerId string, files []Info)
	UpdateOutgoingOfferStatus(offer *Offer, status string, persist bool)
	GetOutgoingOffer(id OfferId) *Offer
	GetOutgoingOffers() Offers
	OutgoingStatus() *Subscriptions

	AddIncomingOffer(peer Peer, offerId string, files []Info)
	UpdateIncomingOfferStatus(offer *Offer, status string, persist bool)
	GetIncomingOffer(id OfferId) *Offer
	GetIncomingOffers() Offers
	IncomingStatus() *Subscriptions

	FilesOffers() *Subscriptions

	CopyFilesFrom(reader io.Reader, offer *Offer) (err error)
	CopyFilesTo(writer io.Writer, offer *Offer) (err error)
}

// Resolver provides access to file by uri.
// Required for platforms where direct access to the file system is restricted.
type Resolver interface {
	Reader(uri string) (io.ReadCloser, error)
	Info(uri string) (files []Info, err error)
}
