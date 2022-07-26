package warpdrive

import "errors"

const Port = "warpdrive"
const CliPort = "wd"

const (

	// local commands
	cmdClose = uint8(iota)

	localPeers
	localSend
	localAccept
	localUpdate
	localSubscribe
	localStatus
	localOffers
	localPing

	// remote commands
	remoteSend
	remoteDownload
)

var errEnded = errors.New("ended")
