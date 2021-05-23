package net

import (
	"context"
	"log"
	"sync"
)

var drivers = make(map[string]Driver)
var mu sync.Mutex

// Dial tries to establish a Conn to provided address
func Dial(ctx context.Context, addr Addr) (Conn, error) {
	log.Println("dial", addr.Network(), addr.String())
	driver := drivers[addr.Network()]

	if driver == nil {
		return nil, ErrUnsupportedNetwork
	}

	return driver.Dial(ctx, addr)
}

// Register sets a driver for a network
func Register(driver Driver) error {
	mu.Lock()
	defer mu.Unlock()

	network := driver.Network()
	if len(network) == 0 {
		return ErrInvalidNetworkName
	}

	if _, found := drivers[network]; found {
		return ErrAlreadyRegistered
	}

	drivers[network] = driver

	return nil
}
