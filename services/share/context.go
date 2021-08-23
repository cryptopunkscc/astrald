package share

import "github.com/cryptopunkscc/astrald/components/shares"

type serviceContext struct {
	shares shares.Shared
}

type requestContext struct {
	serviceContext
}
