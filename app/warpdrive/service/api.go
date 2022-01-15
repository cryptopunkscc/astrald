package warpdrive

import (
	"context"
	astral "github.com/cryptopunkscc/astrald/mod/apphost/client"
	"io"
	"os"
)

// =================== API ===================

type ClientApi interface {
	SenderApi
	RecipientApi
	Sender() SenderApi
	Recipient() RecipientApi
}

type SenderApi interface {
	StatusApi
	// Peers available for receiving an offer.
	Peers() ([]Peer, error)
	// Send files offer for the recipient.
	Send(peerId PeerId, path string) (OfferId, error)
	// Sent offers.
	Sent() (Offers, error)
	Status(id OfferId) (string, error)
}

type RecipientApi interface {
	StatusApi
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

type StatusApi interface {
	// Events subscribes a callback for receiving offers status updates.
	Events() (<-chan Status, error)
}

// ------------------- Peer -------------------

type Peers map[PeerId]*Peer
type PeerId string

type Peer struct {
	Id    PeerId
	Alias string
	Mod   string
}

// ------------------- Offer -------------------

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

type Info struct {
	Path  string
	Size  int64
	IsDir bool
	Perm  os.FileMode
	Mime  string
}

const (
	PEER_MOD_ASK   = ""
	PEER_MOD_TRUST = "trust"
	PEER_MOD_BLOCK = "block"
)

// =================== Dependency ===================

type Resolver interface {
	Reader(uri string) (io.ReadCloser, error)
	Info(uri string) (files []Info, err error)
}

type Config struct {
	context.Context
	astral.Api
	RepositoryDir  string
	RemoteResolver bool
}
