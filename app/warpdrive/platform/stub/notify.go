package stub

import "github.com/cryptopunkscc/astrald/app/warpdrive/api"

var _ api.Notify = &Notifier{}

type Notifier struct{}

func (*Notifier) New(api.Notification)      {}
func (*Notifier) Progress(api.Notification) {}
func (*Notifier) Finish(api.Notification)   {}
