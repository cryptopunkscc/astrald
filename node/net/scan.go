package net

import (
	"context"
	"sync"
)

func Scan(ctx context.Context) (<-chan *Ad, error) {
	output := make(chan *Ad)
	group := sync.WaitGroup{}

	for _, drv := range drivers {
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
