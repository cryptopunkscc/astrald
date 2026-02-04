package nodes

import (
	"math/rand"
	"slices"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/exonet"
	"github.com/cryptopunkscc/astrald/sig"
)

type streamWatcher struct {
	match func(*Stream) bool
	ch    chan *Stream
}

type LinkPool struct {
	peers    *Peers
	mod      *Module
	watchers sig.Set[*streamWatcher]

	// todo: linkers
	// todo:  streams sig.Set[*Stream]
}

func NewLinkPool(mod *Module, peers *Peers) *LinkPool {
	return &LinkPool{
		peers: peers,
		mod:   mod,
	}
}

type LinkFuture <-chan LinkResult

type LinkResult struct {
	Stream *Stream
	Err    error
}

func (pool *LinkPool) subscribeInboundStreams(match func(*Stream) bool) *streamWatcher {
	w := &streamWatcher{
		match: match,
		ch:    make(chan *Stream, 1),
	}

	pool.watchers.Add(w)
	return w
}

func (pool *LinkPool) unsubscribeInboundStreams(w *streamWatcher) {
	pool.watchers.Remove(w)
}

func (pool *LinkPool) processInboundConnection(s *Stream) {
	for _, w := range pool.watchers.Clone() {
		if w.match(s) {
			select {
			case w.ch <- s:
			default:
			}
		}
	}
}

func (pool *LinkPool) RetrieveLink(
	ctx *astral.Context,
	target *astral.Identity,
	opts ...RetrieveLinkOption,
) LinkFuture {

	var o RetrieveLinkOptions
	for _, opt := range opts {
		opt(&o)
	}

	result := make(chan LinkResult, 1)
	match := streamMatcher(target, &o)
	forceNew := o.ForceNew

	if !forceNew {
		streams := pool.peers.streams.Select(match)
		if len(streams) > 0 {
			rand.Shuffle(len(streams), func(i, j int) {
				streams[i], streams[j] = streams[j], streams[i]
			})

			result <- LinkResult{Stream: streams[0]}
			close(result)
			return result
		}
	}

	var endpoints = sig.ArrayToChan(o.Endpoints)
	if len(o.Endpoints) == 0 {
		endpoints = make(chan exonet.Endpoint)
		resolved, err := pool.mod.ResolveEndpoints(ctx, target)
		if err != nil {
			result <- LinkResult{Err: err}
			close(result)
			return result
		}

		endpoints = sig.FilterChan(resolved, endpointFilter(o.IncludeNetworks, o.ExcludeNetworks))
	}

	go func() {
		defer close(result)

		var inboundCh <-chan *Stream
		if !forceNew {
			w := pool.subscribeInboundStreams(match)
			defer pool.unsubscribeInboundStreams(w)
			inboundCh = w.ch
		}

		connectCtx, cancel := ctx.WithCancel()
		defer cancel()

		connectResult := make(chan LinkResult, 1)
		// todo: node linker
		go func() {
			stream, err := pool.peers.connectAtAny(connectCtx, target, endpoints)
			connectResult <- LinkResult{Stream: stream, Err: err}
		}()

		select {
		case <-ctx.Done():
			result <- LinkResult{Err: ctx.Err()}
		case r := <-connectResult:
			result <- r
		case s := <-inboundCh:
			result <- LinkResult{Stream: s}
		}
	}()

	return result
}

func streamMatcher(target *astral.Identity, o *RetrieveLinkOptions) func(*Stream) bool {
	return func(s *Stream) bool {
		if !s.RemoteIdentity().IsEqual(target) {
			return false
		}
		if o == nil {
			return true
		}

		net := s.Network()
		if len(o.ExcludeNetworks) > 0 && slices.Contains(o.ExcludeNetworks, net) {
			return false
		}

		if len(o.IncludeNetworks) > 0 && !slices.Contains(o.IncludeNetworks, net) {
			return false
		}

		return true
	}
}

func endpointFilter(include, exclude []string) func(exonet.Endpoint) bool {
	return func(endpoint exonet.Endpoint) bool {
		net := endpoint.Network()

		if len(exclude) > 0 && slices.Contains(exclude, net) {
			return false
		}

		if len(include) > 0 && !slices.Contains(include, net) {
			return false
		}

		return true
	}
}

// RetrieveLinkOptions controls how RetrieveLink behaves.
type RetrieveLinkOptions struct {
	IncludeNetworks []string
	ExcludeNetworks []string
	Endpoints       []exonet.Endpoint
	ForceNew        bool
}

// RetrieveLinkOption is a functional option for RetrieveLink.
type RetrieveLinkOption func(*RetrieveLinkOptions)

func WithIncludeNetworks(networks ...string) RetrieveLinkOption {
	return func(o *RetrieveLinkOptions) {
		o.IncludeNetworks = append(o.IncludeNetworks, networks...)
	}
}

func WithExcludeNetworks(networks ...string) RetrieveLinkOption {
	return func(o *RetrieveLinkOptions) {
		o.ExcludeNetworks = append(o.ExcludeNetworks, networks...)
	}
}

func WithEndpoints(endpoints ...exonet.Endpoint) RetrieveLinkOption {
	return func(o *RetrieveLinkOptions) {
		o.Endpoints = endpoints
	}
}

func WithForceNew() RetrieveLinkOption {
	return func(o *RetrieveLinkOptions) {
		o.ForceNew = true
	}
}
