package notify

import "log"

var _ Api = Stub{}

type Stub struct{}

func (Stub) Create(channel Channel) error {
	log.Println("Stub create:", channel)
	return nil
}

func (Stub) Notify(notifications ...Notification) error {
	log.Println("Stub notify:", notifications)
	return nil
}
