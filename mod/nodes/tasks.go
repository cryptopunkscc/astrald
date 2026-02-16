package nodes

import "github.com/cryptopunkscc/astrald/mod/scheduler"

type StreamProducerTask interface {
	scheduler.Task
	Result() (info *StreamInfo, err error)
}

type EnsureStreamTask interface {
	StreamProducerTask
}

type CreateStreamTask interface {
	StreamProducerTask
}

type CleanupEndpointsTask interface {
	scheduler.Task
}
