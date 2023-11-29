package sdp

import (
	"context"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/mod/sdp/proto"
)

type API interface {
	AddSource(source Source)
	RemoveSource(source Source)
}

type Source interface {
	Discover(ctx context.Context, caller id.Identity, origin string) ([]ServiceEntry, error)
}

type ServiceEntry proto.ServiceEntry

const DiscoverServiceName = "core.sdp.discover"
const SourceServiceName = "core.sdp.source"
