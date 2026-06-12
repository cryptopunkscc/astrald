package user

import "github.com/cryptopunkscc/astrald/mod/scheduler"

// MaintainLinkTask keeps an active link to a remote peer alive; restarts the
// link when it drops.
type MaintainLinkTask interface {
	scheduler.Task
}
