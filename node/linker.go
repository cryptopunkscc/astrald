package node

import (
	"context"
	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/astral/link"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/node/route"
	"sync"
	"time"
)

type Linker struct {
	localID  id.Identity
	remoteID id.Identity
	routes   map[string]*route.Route
}

func NewLinker(localID id.Identity, remoteID id.Identity, routes map[string]*route.Route) *Linker {
	return &Linker{localID: localID, remoteID: remoteID, routes: routes}
}

func (linker *Linker) Run(ctx context.Context) <-chan *link.Link {
	outCh := make(chan *link.Link)

	go func() {
		defer close(outCh)

		var wg sync.WaitGroup

		for _, netName := range astral.NetworkNames() {
			wg.Add(1)
			go func(netName string) {
				defer wg.Done()

				for lnk := range linker.netLink(netName) {
					outCh <- lnk
				}
			}(netName)
		}

		wg.Wait()
	}()

	return outCh
}

func (linker *Linker) netLink(netName string) <-chan *link.Link {
	outCh := make(chan *link.Link)

	go func() {
		defer close(outCh)

		for {
			r, found := linker.routes[linker.remoteID.PublicKeyHex()]
			if !found {
				time.Sleep(5 * time.Second)
				continue
			}

			for _, addr := range r.Addresses {
				if addr.Network() != netName {
					continue
				}

				lnk, err := astral.Link(linker.localID, linker.remoteID, addr)
				if err != nil {
					// can be very spammy
					// log.Println(err)
					continue
				}
				outCh <- lnk
				<-lnk.WaitClose()
				break
			}

			time.Sleep(5 * time.Second)
		}
	}()

	return outCh
}
