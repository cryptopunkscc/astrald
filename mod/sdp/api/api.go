package sdp

import (
	"context"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/mod/sdp/proto"
	"github.com/cryptopunkscc/astrald/node/modules"
)

const ModuleName = "sdp"

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

func Load(node modules.Node) (API, error) {
	api, ok := node.Modules().Find(ModuleName).(API)
	if !ok {
		return nil, modules.ErrNotFound
	}
	return api, nil
}
