package astral

import (
	"context"
	"github.com/cryptopunkscc/astrald/infra"
	"sync"
)

func (astral *Astral) Discover(ctx context.Context) (<-chan infra.Presence, error) {
	outCh := make(chan infra.Presence)

	var wg sync.WaitGroup

	for network := range astral.Networks() {
		presenceCh, err := network.Discover(ctx)
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
