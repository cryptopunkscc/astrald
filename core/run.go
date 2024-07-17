package core

import (
	"context"
	"fmt"
	"sync"
	"time"
)

// Run starts the node, waits for it to finish and returns an error if any
func (node *Node) Run(ctx context.Context) (err error) {
	ctx, shutdown := context.WithCancel(ctx)

	// Say hello
	node.log.Log("astral node %s (%s) statrting...",
		node.identity,
		node.identity.PublicKeyHex(),
	)

	var wg sync.WaitGroup
	var errCh = make(chan error, 32)

	// run the module manager
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := node.modules.Run(ctx); err != nil {
			errCh <- fmt.Errorf("module manager: %w", err)
		}
	}()

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

	node.startedAt = time.Now()

	// wait for all components to finish
	wg.Wait()

	time.Sleep(100 * time.Millisecond) // wait a little bit of time for all the buffers to flush

	return
}
