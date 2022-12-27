package infra

import "context"

type Network interface {
	// Run the network backend
	Run(ctx context.Context) error
}
