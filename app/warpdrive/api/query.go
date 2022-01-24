package api

const (
	Port = "warpdrive"

	Send   = Port + "/send"
	Accept = Port + "/accept"
	Reject = Port + "/reject"

	SenPeers  = Port + "/sender/peers"
	SenSend   = Port + "/sender/send"
	SenStatus = Port + "/sender/status"
	SenSent   = Port + "/sender/sent"
	SenEvents = Port + "/sender/events"

	RecIncoming = Port + "/recipient/incoming"
	RecReceived = Port + "/recipient/received"
	RecAccept   = Port + "/recipient/accept"
	RecReject   = Port + "/recipient/reject"
	RecUpdate   = Port + "/recipient/update"
	RecEvents   = Port + "/recipient/events"

	CliQuery = "wd"
)
