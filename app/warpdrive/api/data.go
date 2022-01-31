package api

import "os"

type Offers map[OfferId]*Offer
type OfferId string
type Offer struct {
	Status
	Peer       PeerId
	Files      []Info
	CreateTime int64
}
type ResponseCode uint8

const (
	OfferAwaiting = ResponseCode(iota)
	OfferAccepted
)

type Status struct {
	Id     OfferId
	Status string
}

type Peers map[PeerId]*Peer
type PeerId string
type Peer struct {
	Id    PeerId
	Alias string
	Mod   string
}

type Info struct {
	Uri        string
	Size       int64
	IsDir      bool
	Perm       os.FileMode
	Mime       string
	Progress   int64
	UpdateTime int64
}

const (
	StatusAdded     = ""
	StatusAccepted  = "accepted"
	StatusRejected  = "rejected"
	StatusProgress  = "progress"
	StatusFailed    = "failed"
	StatusCompleted = "completed"
	StatusAborted   = "aborted"
)

const (
	PeerModAsk   = ""
	PeerModTrust = "trust"
	PeerModBlock = "block"
)
