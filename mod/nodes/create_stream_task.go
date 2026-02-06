package nodes

import "github.com/cryptopunkscc/astrald/mod/scheduler"

type CreateStreamAction interface {
	scheduler.Task
	Result() (info *StreamInfo, err error)
}
