package net

import (
	"context"
	"sync"
)

// DialFunc describes a function that establishes a Conn to the provided Endpoint
type DialFunc func(context.Context, Endpoint) (Conn, error)

var dials = make(map[string]DialFunc)
var mu sync.Mutex

// Dial tries to establish a Conn to provided Endpoint
func Dial(ctx context.Context, ep Endpoint) (Conn, error) {
	dial := getDial(ep.Net)

	if dial == nil {
		return nil, ErrUnsupportedNetwork
	}

	return dial(ctx, ep)
}

// Register sets a DialFunc for a network
func Register(network string, fun DialFunc) error {
	mu.Lock()
	defer mu.Unlock()

	if _, found := dials[network]; found {
		return ErrAlreadyRegistered
	}

	dials[network] = fun

	return nil
}

func getDial(network string) DialFunc {
	mu.Lock()
	defer mu.Unlock()

	return dials[network]
}
