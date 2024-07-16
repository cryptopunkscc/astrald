package nodes

import (
	"context"
	"errors"
	"github.com/cryptopunkscc/astrald/id"
	"github.com/cryptopunkscc/astrald/mod/exonet"
	"github.com/cryptopunkscc/astrald/mod/nodes"
	"github.com/cryptopunkscc/astrald/mod/nodes/src/muxlink"
	"github.com/cryptopunkscc/astrald/net"
	"github.com/cryptopunkscc/astrald/node"
	"sync"
	"sync/atomic"
)

var _ node.Linker = &Linker{}

type Linker struct {
	*Module
}

func (linker *Linker) Link(ctx context.Context, remoteIdentity id.Identity) (net.Link, error) {
	return linker.LinkOpts(ctx, remoteIdentity, nodes.LinkOpts{})
}

func (linker *Linker) LinkOpts(ctx context.Context, remoteIdentity id.Identity, opts nodes.LinkOpts) (net.Link, error) {
	if linker.node.Identity().IsEqual(remoteIdentity) {
		return nil, errors.New("cannot link to self")
	}

	var endpoints = opts.Endpoints
	if endpoints == nil {
		endpoints = linker.Endpoints(remoteIdentity)
	}

	if len(endpoints) == 0 {
		return nil, errors.New("no endpoints provided")
	}

	var endpointsCh = make(chan exonet.Endpoint, len(endpoints))
	for _, e := range endpoints {
		endpointsCh <- e
	}
	close(endpointsCh)

	var wg sync.WaitGroup
	var linked atomic.Bool
	var res = make(chan net.Link)
	var workers = opts.Workers
	if workers == 0 {
		workers = DefaultWorkerCount
	}

	workerCtx, cancelWorkers := context.WithTimeout(ctx, DefaultTimeout)
	defer cancelWorkers()

	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			for {
				var e exonet.Endpoint
				var ok bool

				select {
				case <-workerCtx.Done():
					return
				case e, ok = <-endpointsCh:
				}

				if !ok {
					return
				}

				conn, err := linker.exonet.Dial(workerCtx, e)
				if err != nil {
					break
				}

				l, err := muxlink.Open(workerCtx, conn, remoteIdentity, linker.node.Identity(), linker.node.LocalRouter())
				if err != nil {
					break
				}

				if !linked.CompareAndSwap(false, true) {
					l.Close()
					return
				}

				res <- l
				return
			}
		}()
	}

	go func() {
		wg.Wait()
		close(res)
	}()

	l := <-res
	if l == nil {
		return nil, errors.New("no endpoint could be reached")
	}

	return l, nil
}
