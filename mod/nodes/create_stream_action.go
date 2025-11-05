package nodes

import "github.com/cryptopunkscc/astrald/mod/scheduler"

type CreateStreamAction interface {
	scheduler.Action
	Result() *StreamInfo
}
