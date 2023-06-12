package node

import (
	"context"
	"fmt"
	"sync"
)

// Run starts the node, waits for it to finish and returns an error if any
func (node *CoreNode) Run(ctx context.Context) (err error) {
	ctx, shutdown := context.WithCancel(ctx)

	// Say hello
	node.log.Log("astral node %s (%s) statrting...",
		node.identity,
		node.identity.PublicKeyHex(),
	)

	var wg sync.WaitGroup
	var errCh = make(chan error, 32)

	// run the infrastructure
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := node.infra.Run(ctx); err != nil {
			errCh <- fmt.Errorf("infrastructure: %w", err)
		}
	}()

	// run the peer manager
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := node.network.Run(ctx); err != nil {
			errCh <- fmt.Errorf("peer manager: %w", err)
		}
	}()

	// run the module manager
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := node.modules.Run(ctx); err != nil {
			errCh <- fmt.Errorf("module manager: %w", err)
		}
	}()

	// peer query worker
	for i := 0; i < 8; i++ {
		wg.Add(1)
		go func() {
			defer wg.Done()
			node.peerQueryWorker(ctx)
		}()
	}

	// event handling
	wg.Add(1)
	go func() {
		defer wg.Done()

		if err := node.handleEvents(ctx); err != nil {
			errCh <- fmt.Errorf("event handler: %w", err)
		}
	}()

	// wait for the context to end or a node error
	go func() {
		select {
		case <-ctx.Done():
		case err = <-errCh:
			shutdown()
		}
	}()

	// wait for all components to finish
	wg.Wait()

	return
}
