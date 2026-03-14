package gateway

import "github.com/cryptopunkscc/astrald/mod/scheduler"

type MaintainBindingTask interface {
	scheduler.Task
}
