package nodes

import (
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
	linkers  sig.Map[string, *NodeLinker]
}

func NewLinkPool(mod *Module, peers *Peers) *LinkPool {
	return &LinkPool{
		peers: peers,
		mod:   mod,
	}
}

type LinkResult struct {
	Stream *Stream
	Err    error
}

func (pool *LinkPool) subscribe(match func(*Stream) bool) *streamWatcher {
	w := &streamWatcher{
		match: match,
		ch:    make(chan *Stream, 1),
	}

	pool.watchers.Add(w)
	return w
}

func (pool *LinkPool) unsubscribe(w *streamWatcher) {
	pool.watchers.Remove(w)
}

func (pool *LinkPool) notifyStream(s *Stream) {
	for _, w := range pool.watchers.Clone() {
		if !w.match(s) {
			continue
		}

		select {
		case w.ch <- s:
		default:
		}
	}
}

func (pool *LinkPool) getOrCreateNodeLinker(target *astral.Identity) *NodeLinker {
	linker := NewNodeLinker(pool.mod, target)

	existing, ok := pool.linkers.Set(target.String(), linker)
	if !ok {
		// already existed, return the existing one
		return existing
	}

	// new linker was inserted — start reader goroutine
	go func() {
		for s := range linker.Produced() {
			pool.notifyStream(s)
		}
	}()

	return linker
}

func (pool *LinkPool) RetrieveLink(
	ctx *astral.Context,
	target *astral.Identity,
	opts ...RetrieveLinkOption,
) <-chan LinkResult {

	var o RetrieveLinkOptions
	for _, opt := range opts {
		opt(&o)
	}

	match := streamMatcher(target, &o)

	if !o.ForceNew {
		streams := pool.peers.streams.Select(match)
		if len(streams) > 0 {
			return sig.ArrayToChan([]LinkResult{{Stream: streams[0]}})
		}
	}

	result := make(chan LinkResult, 1)

	go func() {
		defer close(result)

		w := pool.subscribe(match)
		defer pool.unsubscribe(w)

		linker := pool.getOrCreateNodeLinker(target)
		linker.Activate(ctx, o.LinkConstraints)

		select {
		case <-ctx.Done():
			result <- LinkResult{Err: ctx.Err()}
		case s := <-w.ch:
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
		if len(o.LinkConstraints.ExcludeNetworks) > 0 && slices.Contains(o.LinkConstraints.ExcludeNetworks, net) {
			return false
		}

		if len(o.LinkConstraints.IncludeNetworks) > 0 && !slices.Contains(o.LinkConstraints.IncludeNetworks, net) {
			return false
		}

		return true
	}
}

// RetrieveLinkOptions controls how RetrieveLink behaves.
type RetrieveLinkOptions struct {
	ForceNew        bool
	LinkConstraints LinkConstraints
}

// RetrieveLinkOption is a functional option for RetrieveLink.
type RetrieveLinkOption func(*RetrieveLinkOptions)

func WithIncludeNetworks(networks ...string) RetrieveLinkOption {
	return func(o *RetrieveLinkOptions) {
		o.LinkConstraints.IncludeNetworks = append(o.LinkConstraints.IncludeNetworks, networks...)
	}
}

func WithExcludeNetworks(networks ...string) RetrieveLinkOption {
	return func(o *RetrieveLinkOptions) {
		o.LinkConstraints.ExcludeNetworks = append(o.LinkConstraints.ExcludeNetworks, networks...)
	}
}

func WithEndpoints(endpoints ...exonet.Endpoint) RetrieveLinkOption {
	return func(o *RetrieveLinkOptions) {
		o.LinkConstraints.Endpoints = endpoints
	}
}

func WithForceNew() RetrieveLinkOption {
	return func(o *RetrieveLinkOptions) {
		o.ForceNew = true
	}
}
