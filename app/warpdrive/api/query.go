package api

const (
	Port = "warpdrive"

	QueryPeers     = Port + "/peers"
	QuerySend      = Port + "/send"
	QueryAccept    = Port + "/accept"
	QueryUpdate    = Port + "/update"
	QuerySubscribe = Port + "/subscribe"
	QueryStatus    = Port + "/status"
	QueryOffers    = Port + "/offers"

	QueryOffer = Port + "/remote/offer"
	QueryFiles = Port + "/remote/files"

	QueryCli = "wd"
)
