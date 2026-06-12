package nodes

import "github.com/cryptopunkscc/astrald/mod/scheduler"

// LinkProducerTask is a scheduler task whose Result is valid only after the task completes.
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
