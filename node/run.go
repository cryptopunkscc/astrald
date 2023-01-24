package node

import (
	"context"
	"fmt"
	"github.com/cryptopunkscc/astrald/node/peers"
	"sync"
)

// Run starts the node, waits for it to finish and returns an error if any
func (node *Node) Run(ctx context.Context) (err error) {
	ctx, shutdown := context.WithCancel(ctx)

	// Say hello
	nodeKey := node.identity.PublicKeyHex()
	if node.Alias() != "" {
		log.Log("astral node %s%s%s (%s) statrting...",
			log.Green(),
			node.Alias(),
			log.Reset(),
			log.Em(nodeKey),
		)
	} else {
		log.Log("astral node %s statrting...", log.Em(nodeKey))
	}

	var wg sync.WaitGroup
	var errCh = make(chan error, 32)

	// run the infrastructure
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := node.Infra.Run(ctx); err != nil {
			errCh <- fmt.Errorf("infrastructure: %w", err)
		}
	}()

	// run the peer manager
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := node.Peers.Run(ctx); err != nil {
			errCh <- fmt.Errorf("peer manager: %w", err)
		}
	}()

	// run the module manager
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := node.Modules.Run(ctx); err != nil {
			errCh <- fmt.Errorf("module manager: %w", err)
		}
	}()

	// run presence manager
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := node.Presence.Run(ctx); err != nil {
			errCh <- fmt.Errorf("presence manager: %w", err)
		}
	}()

	// peer query worker
	wg.Add(1)
	go func() {
		defer wg.Done()
		node.peerQueryWorker(ctx)
	}()

	// event handling
	wg.Add(1)
	go func() {
		defer wg.Done()

		if err := node.handleEvents(ctx); err != nil {
			errCh <- fmt.Errorf("event handler: %w", err)
		}
	}()

	// sticky nodes
	wg.Add(1)
	go func() {
		defer wg.Done()
		if err := node.keepStickyNodesLinked(ctx); err != nil {
			errCh <- fmt.Errorf("keeplinked: %w", err)
		}
	}()

	go func() {
		err := node.Events.HandleFunc(ctx, func(e peers.EventPeerLinked) error {
			log.Info("WOW IT WORKS! %s", e.Peer.Identity())
			return nil //errors.New("this is too good")
		})
		if err != nil {
			log.Error("Events.HandleFunc(): %s", err)
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

	if node.Database != nil {
		node.Database.Close()
	}

	return
}
