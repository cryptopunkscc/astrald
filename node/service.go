package node

import (
	"context"
)

type ServiceRunner func(ctx context.Context, node *Node) error

var services = make(map[string]ServiceRunner, 0)

func RegisterService(name string, srv ServiceRunner) error {
	services[name] = srv
	return nil
}
