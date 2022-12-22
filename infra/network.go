package infra

import "context"

type Network interface {
	// Name returns the name of the network
	Name() string

	// Run the network backend
	Run(ctx context.Context) error
}
