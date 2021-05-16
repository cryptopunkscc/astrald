package node

import (
	"context"
	"github.com/cryptopunkscc/astrald/api"
)

var services = make(map[string]Service, 0)

type Service interface {
	Run(ctx context.Context, core api.Core) error
}

func RegisterService(name string, srv Service) error {
	services[name] = srv
	return nil
}
