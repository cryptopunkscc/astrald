package core

import (
	"github.com/cryptopunkscc/astrald/cmd/warpdrived/notify"
	"github.com/cryptopunkscc/astrald/lib/wrapper"
	"github.com/cryptopunkscc/astrald/proto/warpdrive"
	"log"
	"sync"
)

type Component struct {
	Config
	*log.Logger
	*Sys
	*Cache
	*Observers
	*Channel
	wrapper.Api
}

type Config struct {
	RepositoryDir  string
	StorageDir     string
	RemoteResolver bool
	Platform       string
}

type Sys struct {
	Notify notify.Notify
}

type Cache struct {
	*Mutex
	Incoming warpdrive.Offers
	Outgoing warpdrive.Offers
	Peers    warpdrive.Peers
}

type Mutex struct {
	Incoming sync.RWMutex
	Outgoing sync.RWMutex
	Peers    sync.RWMutex
}

type Observers struct {
	IncomingOffers *warpdrive.Subscriptions
	IncomingStatus *warpdrive.Subscriptions
	OutgoingOffers *warpdrive.Subscriptions
	OutgoingStatus *warpdrive.Subscriptions
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
	Stat() warpdrive.OfferStatus
	//Notification related to this update
	Notification() notify.Notification
}

const (
	PlatformDesktop = "desktop"
	PlatformAndroid = "android"
	PlatformDefault = PlatformDesktop
)
