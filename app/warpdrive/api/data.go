package api

import "os"

type Offer struct {
	Status
	Peer  PeerId
	Files []Info
}
type Offers map[OfferId]*Offer
type OfferId string

type Status struct {
	Id     OfferId
	Status string
}

type Info struct {
	Path  string
	Size  int64
	IsDir bool
	Perm  os.FileMode
	Mime  string
}

type Peer struct {
	Id    PeerId
	Alias string
	Mod   string
}
type Peers map[PeerId]*Peer
type PeerId string

const (
	PeerModAsk   = ""
	PeerModTrust = "trust"
	PeerModBlock = "block"
)
