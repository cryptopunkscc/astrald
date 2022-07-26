package stub

import (
	"github.com/cryptopunkscc/astrald/cmd/warpdrived/notify"
)

var _ notify.Notify = Notify

func Notify([]notify.Notification) {}
