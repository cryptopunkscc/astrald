package nodes

import "github.com/cryptopunkscc/astrald/mod/scheduler"

type EnsureStreamTask interface {
	scheduler.Task
	Result() (info *StreamInfo, err error)
}
