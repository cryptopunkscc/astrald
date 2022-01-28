package desktop

import "github.com/cryptopunkscc/astrald/app/warpdrive/api"

var _ api.Notify = Notifier{}

type Notifier struct{}

func (m Notifier) New(n api.Notification) {
	//TODO implement me
}

func (m Notifier) Progress(n api.Notification) {
	//TODO implement me
}

func (m Notifier) Finish(n api.Notification) {
	//TODO implement me
}
