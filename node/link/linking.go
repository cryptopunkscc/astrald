package link

import (
	"context"
	"errors"
	"fmt"
	"github.com/cryptopunkscc/astrald/auth"
	"github.com/cryptopunkscc/astrald/auth/id"
	"github.com/cryptopunkscc/astrald/cslq"
	"github.com/cryptopunkscc/astrald/net"
	"github.com/cryptopunkscc/astrald/node/infra"
	"github.com/cryptopunkscc/astrald/node/tracker"
	"sync"
	"sync/atomic"
)

const DefaultWorkerCount = 8
const featureMux = "mux"
const featureListFormat = "[s][c]c"

type Opts struct {
	Endpoints []net.Endpoint
	Workers   int
}

type Node interface {
	Identity() id.Identity
	Infra() infra.Infra
	Tracker() tracker.Tracker
}

// Open negotiaties a link over the provided conn as the active party
func Open(
	ctx context.Context,
	conn net.Conn,
	remoteID id.Identity,
	localID id.Identity,
) (link *CoreLink, err error) {
	defer func() {
		if err != nil {
			conn.Close()
		}
	}()

	secureConn, err := auth.HandshakeOutbound(ctx, conn, remoteID, localID)
	if err != nil {
		return
	}

	var linkFeatures []string

	err = cslq.Decode(secureConn, featureListFormat, &linkFeatures)
	if err != nil {
		return
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

	err = cslq.Encode(secureConn, "[c]c", featureMux)
	if err != nil {
		return
	}

	var errCode int
	err = cslq.Decode(secureConn, "c", &errCode)
	if errCode != 0 {
		err = errors.New("link feature negotation error")
		return
	}

	return NewCoreLink(secureConn), nil
}

// Accept negotiaties a link over the provided conn as the passive party
func Accept(
	ctx context.Context,
	conn net.Conn,
	localID id.Identity,
) (link *CoreLink, err error) {
	defer func() {
		if err != nil {
			conn.Close()
		}
	}()

	secureConn, err := auth.HandshakeInbound(ctx, conn, localID)
	if err != nil {
		return
	}

	var linkFeatures = []string{featureMux}

	err = cslq.Encode(secureConn, featureListFormat, linkFeatures)
	if err != nil {
		return
	}

	for {
		var feature string
		err = cslq.Decode(secureConn, "[c]c", &feature)
		if err != nil {
			return
		}

		switch feature {
		case featureMux:
			cslq.Encode(secureConn, "c", 0)
			return NewCoreLink(secureConn), nil

		default:
			cslq.Encode(secureConn, "c", 1)
			return nil, fmt.Errorf("remote party (%s from %s) requested an invalid feature: %s",
				secureConn.RemoteIdentity(),
				secureConn.RemoteEndpoint(),
				feature,
			)
		}
	}
}

// MakeLink tries to establish a new link from the node to the provided identity. It does not automatically add
// the new link to node's network. By default, it will try to reach the node using all endpoints from node's tracker.
// This can be overriden in opts. MakeLink will spawn DefaultWorkerCount concurrent workers, unless overriden by
// opts.
func MakeLink(
	ctx context.Context,
	node Node,
	remoteIdentity id.Identity,
	opts Opts,
) (*CoreLink, error) {
	var localIdentity = node.Identity()

	if localIdentity.PrivateKey() == nil {
		return nil, errors.New("private key missing")
	}

	if localIdentity.IsEqual(remoteIdentity) {
		return nil, errors.New("cannot link to self")
	}

	var endpoints = opts.Endpoints
	if endpoints == nil {
		endpoints, _ = node.Tracker().EndpointsByIdentity(remoteIdentity)
	}

	if len(endpoints) == 0 {
		return nil, errors.New("no endpoints provided")
	}

	var endpointsCh = make(chan net.Endpoint, len(endpoints))
	for _, e := range endpoints {
		endpointsCh <- e
	}
	close(endpointsCh)

	var wg sync.WaitGroup
	var linked atomic.Bool
	var res = make(chan *CoreLink)
	var workers = opts.Workers
	if workers == 0 {
		workers = DefaultWorkerCount
	}

	workerCtx, cancelWorkers := context.WithCancel(ctx)
	defer cancelWorkers()

	wg.Add(workers)
	for i := 0; i < workers; i++ {
		go func() {
			defer wg.Done()
			for {
				var e net.Endpoint
				var ok bool

				select {
				case <-workerCtx.Done():
					return
				case e, ok = <-endpointsCh:
				}

				if !ok {
					return
				}

				conn, err := node.Infra().Dial(workerCtx, e)
				if err != nil {
					break
				}

				link, err := Open(workerCtx, conn, remoteIdentity, localIdentity)
				if err != nil {
					break
				}

				if !linked.CompareAndSwap(false, true) {
					link.Close()
					return
				}

				res <- link
				return
			}
		}()
	}

	go func() {
		wg.Wait()
		close(res)
	}()

	link := <-res
	if link == nil {
		return nil, errors.New("no endpoint could be reached")
	}

	return link, nil
}
