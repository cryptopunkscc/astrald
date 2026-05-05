package nodes

import "github.com/cryptopunkscc/astrald/mod/scheduler"

type LinkProducerTask interface {
	scheduler.Task
	Result() (info *LinkInfo, err error)
}

type EnsureLinkTask interface {
	LinkProducerTask
}

type CreateLinkTask interface {
	LinkProducerTask
}

type CleanupEndpointsTask interface {
	scheduler.Task
}
