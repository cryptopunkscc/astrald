package nodes

import "github.com/cryptopunkscc/astrald/mod/scheduler"

type CreateStreamTask interface {
	scheduler.Task
	Result() (info *StreamInfo, err error)
}
