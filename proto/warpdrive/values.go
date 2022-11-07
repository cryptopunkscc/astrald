package warpdrive

import "errors"

const (
	Port    = "warpdrive"
	PortCli = "wd"
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
	remoteSend = uint8(iota) + 100
	remoteDownload
)

// info commands
const (
	infoPing = uint8(iota) + 200
)

var errEnded = errors.New("ended")
