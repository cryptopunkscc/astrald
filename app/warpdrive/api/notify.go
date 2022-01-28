package api

type Notifications struct {
	Notify chan<- Notification
}

type Notification struct {
	Offer
	*Info
	Incoming bool
}

type Notify interface {
	New(Notification)
	Progress(Notification)
	Finish(Notification)
}
