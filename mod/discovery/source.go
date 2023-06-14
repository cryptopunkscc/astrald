package discovery

import (
	"context"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/cslq"
	"github.com/cryptopunkscc/astrald/mod/discovery/rpc"
	"github.com/cryptopunkscc/astrald/node/services"
	"github.com/cryptopunkscc/astrald/query"
)

type Source interface {
	Discover(ctx context.Context, caller id.Identity, origin string) ([]ServiceEntry, error)
}

var _ Source = &ServiceSource{}

type ServiceSource struct {
	identity id.Identity
	services services.Services
	service  string
}

func (src *ServiceSource) String() string {
	return src.service
}

func (src *ServiceSource) Discover(ctx context.Context, caller id.Identity, origin string) ([]ServiceEntry, error) {
	conn, err := query.Run(ctx,
		src.services,
		query.New(caller, src.identity, src.service),
	)
	if err != nil {
		return nil, err
	}

	var list = make([]ServiceEntry, 0)

	for err == nil {
		err = cslq.Invoke(conn, func(msg rpc.ServiceEntry) error {
			msg.Identity = conn.RemoteIdentity()
			list = append(list, ServiceEntry(msg))
			return nil
		})
	}

	return list, nil
}
