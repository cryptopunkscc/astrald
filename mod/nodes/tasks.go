package nodes

import "github.com/cryptopunkscc/astrald/mod/scheduler"

type LinkProducerTask interface {
	scheduler.Task
	Result() (info *LinkInfo, err error)
}

type StreamProducerTask = LinkProducerTask

type EnsureLinkTask interface {
	LinkProducerTask
}

type EnsureStreamTask = EnsureLinkTask

type CreateLinkTask interface {
	LinkProducerTask
}

type CreateStreamTask = CreateLinkTask

type CleanupEndpointsTask interface {
	scheduler.Task
}
