package nodes

import (
	"context"
	"errors"
	"fmt"
	"github.com/cryptopunkscc/astrald/cslq"
	"github.com/cryptopunkscc/astrald/id"
	"github.com/cryptopunkscc/astrald/mod/exonet"
	"github.com/cryptopunkscc/astrald/mod/nodes/src/noise"
	"github.com/cryptopunkscc/astrald/sig"
	"io"
	"slices"
	"sync"
	"sync/atomic"
)

func (mod *Module) Connect(ctx context.Context, remoteID id.Identity, conn exonet.Conn) (link io.Closer, err error) {
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

	if slices.Contains(linkFeatures, featureMux2) {
		err = cslq.Encode(aconn, "[c]c", featureMux2)
		if err != nil {
			return nil, fmt.Errorf("write: %w", err)
		}

		var errCode int
		err = cslq.Decode(aconn, "c", &errCode)
		if errCode != 0 {
			return nil, errors.New("link feature negotation error")
		}

		mod.addStream(newStream(aconn, true))

		return nil, err
	}

	return nil, errors.New("no supported link types found")
}

func (mod *Module) Accept(ctx context.Context, conn exonet.Conn) (err error) {
	defer func() {
		if err != nil {
			conn.Close()
		}
	}()

	aconn, err := noise.HandshakeInbound(ctx, conn, mod.node.Identity())
	if err != nil {
		return
	}

	var linkFeatures = []string{featureMux2}

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
		case featureMux2:
			err = cslq.Encode(aconn, "c", 0)
			if err == nil {
				mod.addStream(newStream(aconn, false))
			}

			return

		default:
			cslq.Encode(aconn, "c", 1)
			return fmt.Errorf("remote party (%s from %s) requested an invalid feature: %s",
				aconn.RemoteIdentity(),
				aconn.RemoteEndpoint(),
				feature,
			)
		}
	}
}

func (mod *Module) connectAt(ctx context.Context, remoteIdentity id.Identity, e exonet.Endpoint) error {
	conn, err := mod.Exonet.Dial(ctx, e)
	if err != nil {
		return err
	}

	_, err = mod.Connect(ctx, remoteIdentity, conn)
	if err != nil {
		return err
	}

	return nil
}

func (mod *Module) connectAny(ctx context.Context, remoteIdentity id.Identity, endpoints []exonet.Endpoint) error {
	var queue = sig.ArrayToChan(endpoints)

	if len(queue) == 0 {
		return errors.New("no endpoints provided")
	}

	var wg sync.WaitGroup
	var success atomic.Bool
	var workers = DefaultWorkerCount

	wctx, cancel := context.WithCancel(ctx)
	defer cancel()

	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			for {
				var e exonet.Endpoint
				var ok bool

				select {
				case <-wctx.Done():
					return
				case e, ok = <-queue:
					if !ok {
						return
					}
				}

				err := mod.connectAt(wctx, remoteIdentity, e)
				if err == nil {
					success.Store(true)
					cancel()
					return
				}
			}
		}()
	}

	go func() {
		wg.Wait()
		cancel()
	}()

	<-wctx.Done()
	if success.Load() {
		return nil
	}

	return errors.New("no endpoint could be reached")
}

func (mod *Module) ensureConnected(ctx context.Context, remoteIdentity id.Identity) error {
	if mod.isLinked(remoteIdentity) {
		return nil
	}

	return mod.connectAny(ctx, remoteIdentity, mod.Endpoints(remoteIdentity))
}
