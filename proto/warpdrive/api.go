package warpdrive

import (
	"io"
	"os"
)

type Service interface {
	Incoming() OfferService
	Outgoing() OfferService
	Peer() PeerService
	File() FileService
}

type OfferService interface {
	List() (offers []Offer)
	Get(id OfferId) *Offer
	Add(offerId OfferId, files []Info, peerId PeerId) *Offer
	Accept(offer *Offer)
	Finish(offer *Offer, err error)
	OfferSubscriptions() *Subscriptions
	StatusSubscriptions() *Subscriptions
}

type PeerService interface {
	Fetch()
	Update(peerId string, attr string, value string)
	Get(id PeerId) Peer
	List() (peers []Peer)
}

type FileService interface {
	Info(uri string) (files []Info, err error)
	Copy(offer *Offer) CopyOffer
}

type CopyOffer interface {
	From(reader io.Reader) (err error)
	To(writer io.Writer) (err error)
}

type Offers map[OfferId]*Offer
type OfferId string
type Offer struct {
	OfferStatus
	// Create time
	Create int64
	// Peer unique identifier
	Peer PeerId
	// Files info
	Files []Info
}

const (
	OfferAwaiting = false
	OfferAccepted = true
)

type OfferStatus struct {
	// Id the unique offer identifier.
	Id OfferId
	// In marks if offer is incoming or outgoing.
	In bool
	// Status of the offer
	Status string
	// Index of transferred files. If transfer is not started the index is equal -1.
	Index int
	// Progress of specific file transfer
	Progress int64
	// Update timestamp in milliseconds
	Update int64
}

func (offer Offer) IsOngoing() bool {
	return offer.Index > -1 && offer.Index < len(offer.Files)
}

type Peers map[PeerId]*Peer

type PeerId string
type Peer struct {
	Id    PeerId
	Alias string
	Mod   string
}
type Info struct {
	Uri   string
	Path  string
	Size  int64
	IsDir bool
	Perm  os.FileMode
	Mime  string
	Name  string
}

const (
	StatusAwaiting  = "awaiting"
	StatusAccepted  = "accepted"
	StatusRejected  = "rejected"
	StatusUpdated   = "updated"
	StatusCompleted = "completed"
	StatusFailed    = "failed"
)

const (
	PeerModAsk   = ""
	PeerModTrust = "trust"
	PeerModBlock = "block"
)

type Filter string

const (
	FilterAll = Filter("all")
	FilterIn  = Filter("in")
	FilterOut = Filter("out")
)
