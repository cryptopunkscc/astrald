package linking

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/infra"
	"github.com/cryptopunkscc/astrald/link"
	"github.com/cryptopunkscc/astrald/node/contacts"
	"github.com/cryptopunkscc/astrald/node/peer"
	"sync"
	"time"
)

const DefaultConcurrency = 8

type LinkHandlerFunc func(*link.Link) error

type NetworkOptimizer struct {
	localID  id.Identity
	remoteID id.Identity
	network  string

	contacts *contacts.Manager
	peers    *peer.Manager
	dialer   infra.Dialer

	handleLink LinkHandlerFunc

	mu     sync.Mutex
	cancel context.CancelFunc
}

func NewNetworkOptimizer(
	localID id.Identity,
	remoteID id.Identity,
	network string,
	contacts *contacts.Manager,
	peers *peer.Manager,
	dialer infra.Dialer,
	linkHandler LinkHandlerFunc,
) *NetworkOptimizer {
	return &NetworkOptimizer{
		localID:    localID,
		remoteID:   remoteID,
		network:    network,
		contacts:   contacts,
		peers:      peers,
		dialer:     dialer,
		handleLink: linkHandler,
	}
}

func (o *NetworkOptimizer) Start() {
	o.mu.Lock()
	defer o.mu.Unlock()

	if o.cancel != nil {
		return
	}

	var ctx context.Context
	ctx, o.cancel = context.WithCancel(context.Background())
	go o.optimize(ctx)
}

func (o *NetworkOptimizer) Stop() {
	o.mu.Lock()
	defer o.mu.Unlock()

	if o.cancel == nil {
		return
	}
	o.cancel()
	o.cancel = nil
}

func (o *NetworkOptimizer) establishNewLink(ctx context.Context) *link.Link {
	addrCh := make(chan *contacts.Addr, 1024)
	connCh := make(chan infra.Conn, 1)

	// follow contact's addresses
	go func() {
		defer close(addrCh)
		contact := o.contacts.Find(o.remoteID, false)
		if contact == nil {
			return // contact does not exist
		}

		for addr := range contact.SubscribeAddr(ctx.Done(), true) {
			if addr.Network() == o.network {
				addrCh <- addr
			}
		}
	}()

	// produce connections
	var wg sync.WaitGroup
	wg.Add(DefaultConcurrency)
	for i := 0; i < DefaultConcurrency; i++ {
		go func() {
			defer wg.Done()
			for addr := range addrCh {
				conn, err := o.dialer.Dial(ctx, addr.Addr)
				if errors.Is(err, context.Canceled) {
					return
				}
				if err != nil {
					select {
					case <-time.After(time.Second):
						addrCh <- addr
					case <-ctx.Done():
						return
					}
					continue
				}
				connCh <- conn
			}
		}()
	}

	// close conn channel after all dialers are done
	go func() {
		wg.Wait()
		close(connCh)
	}()

	return LinkFirst(ctx, o.localID, o.remoteID, connCh)
}

func (o *NetworkOptimizer) optimize(ctx context.Context) error {
	for {
		var wg sync.WaitGroup

		// if peer is already connected, wait until all links on the network go down
		if peer := o.peers.Find(o.remoteID); peer != nil {
			for link := range peer.Links() {
				if link.Network() != o.network {
					continue
				}
				wg.Add(1)
				go func() {
					defer wg.Done()
					select {
					case <-link.Wait():
					case <-ctx.Done():
					}
				}()
			}
			wg.Wait()
		}

		// check context
		select {
		case <-ctx.Done():
			return nil
		default:
		}

		// establish a link
		if link := o.establishNewLink(ctx); link != nil {
			if err := o.handleLink(link); err != nil {
				return err
			}

			select {
			case <-link.Wait():
			case <-ctx.Done():
				return nil
			}
		} else {
			select {
			case <-time.After(time.Second):
			case <-ctx.Done():
				return nil
			}
		}
	}

}
