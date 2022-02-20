package api

import "os"

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

type Peers map[PeerId]*Peer

type PeerId string
type Peer struct {
	Id    PeerId
	Alias string
	Mod   string
}
type Info struct {
	Uri   string
	Size  int64
	IsDir bool
	Perm  os.FileMode
	Mime  string
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
