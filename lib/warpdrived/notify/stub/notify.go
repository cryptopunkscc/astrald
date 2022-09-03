package stub

import (
	"github.com/cryptopunkscc/astrald/lib/warpdrived/notify"
)

var _ notify.Notify = Notify

func Notify([]notify.Notification) {}
