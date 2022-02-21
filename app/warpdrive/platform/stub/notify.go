package stub

import "github.com/cryptopunkscc/astrald/app/warpdrive/api"

var _ api.Notify = Notify

func Notify([]api.Notification) {}
