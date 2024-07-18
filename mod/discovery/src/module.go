package discovery

import (
	"context"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/core/assets"
	"github.com/cryptopunkscc/astrald/cslq"
	"github.com/cryptopunkscc/astrald/events"
	"github.com/cryptopunkscc/astrald/id"
	"github.com/cryptopunkscc/astrald/lib/routers"
	"github.com/cryptopunkscc/astrald/log"
	"github.com/cryptopunkscc/astrald/mod/dir"
	"github.com/cryptopunkscc/astrald/mod/discovery"
	"github.com/cryptopunkscc/astrald/mod/discovery/proto"
	"github.com/cryptopunkscc/astrald/sig"
	"github.com/cryptopunkscc/astrald/tasks"
)

var _ discovery.Module = &Module{}

type Module struct {
	*routers.PathRouter
	node     astral.Node
	events   events.Queue
	config   Config
	assets   assets.Assets
	log      *log.Logger
	services sig.Set[discovery.ServiceDiscoverer]
	data     sig.Set[discovery.DataDiscoverer]
	ctx      context.Context
	dir      dir.Module
}

func (mod *Module) Run(ctx context.Context) error {
	mod.ctx = ctx

	return tasks.Group(
		&DiscoveryService{Module: mod},
		&EventHandler{Module: mod},
	).Run(ctx)
}

func (mod *Module) AddDataDiscoverer(d discovery.DataDiscoverer) error {
	return mod.data.Add(d)
}

func (mod *Module) RemoveDataDiscoverer(d discovery.DataDiscoverer) error {
	return mod.data.Remove(d)
}

func (mod *Module) AddServiceDiscoverer(d discovery.ServiceDiscoverer) error {
	return mod.services.Add(d)
}

func (mod *Module) RemoveServiceDiscoverer(d discovery.ServiceDiscoverer) error {
	return mod.services.Remove(d)
}

func (mod *Module) DiscoverLocal(ctx context.Context, caller id.Identity, origin string) (*discovery.Info, error) {
	var info = &discovery.Info{}

	for _, d := range mod.data.Clone() {
		dataItems, err := d.DiscoverData(ctx, caller, origin)
		if err != nil {
			continue
		}

		for _, data := range dataItems {
			info.Data = append(info.Data, discovery.Data{Bytes: data})
		}
	}

	for _, d := range mod.services.Clone() {
		services, err := d.DiscoverServices(ctx, caller, origin)
		if err != nil {
			continue
		}

		info.Services = append(info.Services, services...)
	}

	return info, nil
}

func (mod *Module) DiscoverRemote(ctx context.Context, remoteID id.Identity, callerID id.Identity) (*discovery.Info, error) {
	if callerID.IsZero() {
		callerID = mod.node.Identity()
	}

	conn, err := astral.Route(ctx,
		mod.node.Router(),
		astral.NewQuery(callerID, remoteID, discovery.DiscoverServiceName),
	)
	if err != nil {
		return nil, err
	}
	defer conn.Close()

	var info discovery.Info
	var pInfo proto.Info

	err = cslq.Decode(conn, "v", &pInfo)
	if err != nil {
		return nil, err
	}

	for _, c := range pInfo.Data {
		info.Data = append(info.Data, discovery.Data{Bytes: c.Bytes})
	}

	for _, s := range pInfo.Services {
		info.Services = append(info.Services, discovery.Service{
			Identity: s.Identity,
			Name:     s.Name,
			Type:     s.Type,
			Extra:    s.Extra,
		})
	}

	return &info, nil
}
