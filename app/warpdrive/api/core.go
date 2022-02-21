package api

import (
	"log"
	"sync"
)

type Core struct {
	Config
	*log.Logger
	*Sys
	*Cache
	*Observers
	Channel *Channel
}

type Config struct {
	RepositoryDir  string
	StorageDir     string
	RemoteResolver bool
	Platform       string
}

type Sys struct {
	Notify Notify
}

type Cache struct {
	*Mutex
	Incoming Offers
	Outgoing Offers
	Peers    Peers
}

type Mutex struct {
	Incoming sync.RWMutex
	Outgoing sync.RWMutex
	Peers    sync.RWMutex
}

type Observers struct {
	IncomingOffers *Subscriptions
	IncomingStatus *Subscriptions
	OutgoingOffers *Subscriptions
	OutgoingStatus *Subscriptions
}

type Channel struct {
	Offers chan<- OfferUpdate
}

type OfferUpdate interface {
	//Cache updated offer in memory
	Cache()
	//Save  updated offer in persistent storage
	Save()
	//Forward update to listeners
	Forward()
	//A Stat returns related OfferStatus
	Stat() OfferStatus
	//Notification related to this update
	Notification() Notification
}

const (
	PlatformDesktop = "desktop"
	PlatformAndroid = "android"
	PlatformDefault = PlatformDesktop
)
