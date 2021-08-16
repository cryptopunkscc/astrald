package net

import (
	"context"
	"github.com/cryptopunkscc/astrald/node/auth/id"
	"sync"
)

type UnicastNetwork interface {
	Dial(ctx context.Context, addr Addr) (Conn, error)
	Listen(ctx context.Context) (<-chan Conn, error)
}

type BroadcastNetwork interface {
	Advertise(ctx context.Context, id string) error
	Scan(ctx context.Context) (<-chan *Ad, error)
}

type Ad struct {
	Identity id.Identity
	Addr     Addr
}

var unicastNets = make(map[string]UnicastNetwork)
var broadcastNets = make(map[string]BroadcastNetwork)
var mu sync.Mutex

func AddUnicastNetwork(name string, driver UnicastNetwork) error {
	mu.Lock()
	defer mu.Unlock()

	if len(name) == 0 {
		return ErrInvalidNetworkName
	}
	if _, found := unicastNets[name]; found {
		return ErrAlreadyRegistered
	}

	unicastNets[name] = driver

	return nil
}

func AddBroadcastNetwork(name string, driver BroadcastNetwork) error {
	mu.Lock()
	defer mu.Unlock()

	if len(name) == 0 {
		return ErrInvalidNetworkName
	}
	if _, found := broadcastNets[name]; found {
		return ErrAlreadyRegistered
	}

	broadcastNets[name] = driver

	return nil
}
