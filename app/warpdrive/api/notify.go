package api

type Notify func([]Notification)

type Notification struct {
	Peer
	Offer
	*Info
}
