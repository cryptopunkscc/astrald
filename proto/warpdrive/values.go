package warpdrive

import "errors"

const (
	PortCli    = "wd"
	PortLocal  = "warpdrive-local"
	PortRemote = "warpdrive-remote"
	PortInfo   = "warpdrive-info"
)

// common commands
const (
	cmdClose = uint8(iota) + 0xFF
)

// local commands
const (
	localListPeers = uint8(iota) + 1
	localCreateOffer
	localAcceptOffer
	localListOffers
	localListenOffers
	localListenStatus
	localUpdatePeer
)

// remote commands
const (
	remoteSend = uint8(iota) + 1
	remoteDownload
)

// info commands
const (
	infoPing = uint8(iota) + 1
)

var errEnded = errors.New("ended")
