package network

import (
	"context"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/link"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/infra/inet"
	"github.com/cryptopunkscc/astrald/infra/tor"
	"log"
)

type Network struct {
	*View
	requests chan link.Request
	config   Config
	inet     *inet.Inet
	tor      *tor.Tor
}

func NewNetwork(config Config) *Network {
	var err error
	n := &Network{
		config:   config,
		requests: make(chan link.Request),
		View:     NewView(),
	}

	// Configure internet
	n.inet = inet.New()
	if config.ExternalAddr != "" {
		err := n.inet.AddExternalAddr(config.ExternalAddr)
		if err != nil {
			log.Println("config error: external ip:", err)
		}
	}
	err = astral.AddNetwork(n.inet)
	if err != nil {
		panic(err)
	}

	// Configure tor
	n.tor = tor.New()
	err = astral.AddNetwork(n.tor)
	if err != nil {
		panic(err)
	}

	return n
}

func (n *Network) Run(ctx context.Context, localID id.Identity) (<-chan link.Request, <-chan error) {
	errCh := make(chan error)

	go func() {
		defer close(errCh)
		listenCtx, listenCancel := context.WithCancel(ctx)
		defer listenCancel()

		listen, listenErrCh := astral.Listen(listenCtx, localID)

		for {
			select {
			case link := <-listen:
				if err := n.AddLink(link); err != nil {
					log.Println("error adding link:", err)
					_ = link.Close()
				}
			case errCh <- <-listenErrCh:
				return
			case <-ctx.Done():
				errCh <- ctx.Err()
				return
			}
		}
	}()

	return n.requests, errCh
}

func (n *Network) AddLink(link *link.Link) error {
	err := n.View.AddLink(link)
	if err != nil {
		return err
	}

	go func() {
		for req := range link.Requests() {
			n.requests <- req
		}
	}()

	return nil
}

func (n *Network) Requests() chan link.Request {
	return n.requests
}
