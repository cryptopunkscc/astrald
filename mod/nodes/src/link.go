package nodes

import (
	"context"
	"errors"
	"fmt"
	"github.com/cryptopunkscc/astrald/cslq"
	"github.com/cryptopunkscc/astrald/id"
	"github.com/cryptopunkscc/astrald/mod/exonet"
	"github.com/cryptopunkscc/astrald/mod/nodes"
	"github.com/cryptopunkscc/astrald/mod/nodes/src/muxlink"
	"github.com/cryptopunkscc/astrald/mod/nodes/src/noise"
	"sync"
	"sync/atomic"
)

func (mod *Module) AcceptLink(ctx context.Context, conn exonet.Conn) (link nodes.Link, err error) {
	defer func() {
		if err != nil {
			conn.Close()
		}
	}()

	aconn, err := noise.HandshakeInbound(ctx, conn, mod.node.Identity())
	if err != nil {
		return
	}

	var linkFeatures = []string{featureMux}

	err = cslq.Encode(aconn, "[s][c]c", linkFeatures)
	if err != nil {
		return
	}

	for {
		var feature string
		err = cslq.Decode(aconn, "[c]c", &feature)
		if err != nil {
			return
		}

		switch feature {
		case featureMux:
			cslq.Encode(aconn, "c", 0)
			link = muxlink.NewLink(aconn, mod.node.Router())

			err = mod.addLink(link)
			if err != nil {
				link.Close()
			}

			return

		default:
			cslq.Encode(aconn, "c", 1)
			return nil, fmt.Errorf("remote party (%s from %s) requested an invalid feature: %s",
				aconn.RemoteIdentity(),
				aconn.RemoteEndpoint(),
				feature,
			)
		}
	}
}

func (mod *Module) InitLink(ctx context.Context, conn exonet.Conn, remoteID id.Identity) (link nodes.Link, err error) {
	defer func() {
		if err != nil {
			conn.Close()
		}
	}()

	aconn, err := noise.HandshakeOutbound(ctx, conn, remoteID, mod.node.Identity())
	if err != nil {
		return nil, fmt.Errorf("outbound handshake: %w", err)
	}

	var linkFeatures []string

	err = cslq.Decode(aconn, "[s][c]c", &linkFeatures)
	if err != nil {
		return nil, fmt.Errorf("read features: %w", err)
	}

	var muxFound bool
	for _, f := range linkFeatures {
		if f == featureMux {
			muxFound = true
		}
	}
	if !muxFound {
		return nil, errors.New("remote party does not support mux")
	}

	err = cslq.Encode(aconn, "[c]c", featureMux)
	if err != nil {
		return nil, fmt.Errorf("write mux: %w", err)
	}

	var errCode int
	err = cslq.Decode(aconn, "c", &errCode)
	if errCode != 0 {
		return nil, errors.New("link feature negotation error")
	}

	link = muxlink.NewLink(aconn, mod.node.Router())

	err = mod.addLink(link)
	if err != nil {
		link.Close()
	}

	return
}

func (mod *Module) Link(ctx context.Context, remoteIdentity id.Identity, opts nodes.LinkOpts) (nodes.Link, error) {
	if mod.node.Identity().IsEqual(remoteIdentity) {
		return nil, errors.New("cannot link to self")
	}

	var endpoints = opts.Endpoints
	if endpoints == nil {
		endpoints = mod.Endpoints(remoteIdentity)
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
	var res = make(chan nodes.Link)
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

				conn, err := mod.exonet.Dial(workerCtx, e)
				if err != nil {
					break
				}

				l, err := mod.InitLink(workerCtx, conn, remoteIdentity)
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
		//TODO: return context error if there was a context error
		return nil, errors.New("no endpoint could be reached")
	}

	return l, nil
}

func (mod *Module) addLink(link nodes.Link) error {
	link.SetLocalRouter(mod.node.Router())
	err := mod.links.Add(link)
	if err != nil {
		return err
	}

	mod.log.Logv(1, "added link with %v (%s)", link.RemoteIdentity(), Network(link))

	go func() {
		link.Run(context.Background())
		mod.links.Remove(link)
		mod.log.Logv(1, "removed link with %v (%s)", link.RemoteIdentity(), Network(link))
	}()

	return nil
}
