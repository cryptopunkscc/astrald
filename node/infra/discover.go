package infra

import (
	"context"
	"github.com/cryptopunkscc/astrald/infra"
	"sync"
)

var _ infra.Discoverer = &Infra{}

func (i *Infra) Discover(ctx context.Context) (<-chan infra.Presence, error) {
	outCh := make(chan infra.Presence)

	var wg sync.WaitGroup

	for network := range i.Networks() {
		discoverer, ok := network.(infra.Discoverer)
		if !ok {
			continue
		}

		presenceCh, err := discoverer.Discover(ctx)
		if err != nil {
			continue
		}

		wg.Add(1)
		go func() {
			defer wg.Done()
			for presence := range presenceCh {
				outCh <- presence
			}
		}()
	}

	go func() {
		// wait for all networks to finish
		wg.Wait()
		close(outCh)
	}()

	return outCh, nil
}
