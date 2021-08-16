package net

import (
	"context"
	"sync"
)

func Advertise(ctx context.Context, id string) error {
	for _, drv := range broadcastNets {
		drv.Advertise(ctx, id)
	}
	return nil
}

func Scan(ctx context.Context) (<-chan *Ad, error) {
	output := make(chan *Ad)
	group := sync.WaitGroup{}

	for _, drv := range broadcastNets {
		ads, err := drv.Scan(ctx)
		if err != nil {
			continue
		}

		group.Add(1)
		go func() {
			defer group.Done()
			for ad := range ads {
				output <- ad
			}
		}()
	}

	go func() {
		group.Wait()
		close(output)
	}()

	return output, nil
}
