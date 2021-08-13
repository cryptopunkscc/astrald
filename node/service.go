package node

import (
	"context"
	"github.com/cryptopunkscc/astrald/api"
)

type ServiceRunner func(ctx context.Context, core api.Core) error

var services = make(map[string]ServiceRunner, 0)

func RegisterService(name string, srv ServiceRunner) error {
	services[name] = srv
	return nil
}
