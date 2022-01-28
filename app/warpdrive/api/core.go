package api

import (
	"log"
	"sync"
)

type Core struct {
	Config
	*log.Logger
	*Cache
	*Observers
	*Notifications
}

type Config struct {
	RepositoryDir  string
	StorageDir     string
	RemoteResolver bool
	Platform       string
}

type Cache struct {
	*Mutex
	Incoming Offers
	Outgoing Offers
	Peers    Peers
}

type Mutex struct {
	Incoming sync.Mutex
	Outgoing sync.Mutex
	Peers    sync.Mutex
}

type Observers struct {
	FilesOffers    *Subscriptions
	IncomingStatus *Subscriptions
	OutgoingStatus *Subscriptions
}

const (
	PlatformDesktop = "desktop"
	PlatformAndroid = "android"
	PlatformDefault = PlatformDesktop
)
