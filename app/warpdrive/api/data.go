package api

import "os"

type Offers map[OfferId]*Offer
type OfferId string
type Offer struct {
	Status
	Peer  PeerId
	Files []Info
}
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
	Path  string
	Size  int64
	IsDir bool
	Perm  os.FileMode
	Mime  string
}

const (
	PeerModAsk   = ""
	PeerModTrust = "trust"
	PeerModBlock = "block"
)
