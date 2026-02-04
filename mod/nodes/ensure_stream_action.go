package nodes

import "github.com/cryptopunkscc/astrald/mod/scheduler"

type EnsureStreamAction interface {
	scheduler.Task
	Result() (info *StreamInfo, err error)
}
