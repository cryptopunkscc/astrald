package nodes

import (
	"slices"

	"github.com/cryptopunkscc/astrald/astral"
	"github.com/cryptopunkscc/astrald/mod/nodes"
	"github.com/cryptopunkscc/astrald/sig"
)

type streamWatcher struct {
	match func(*Stream, *string) bool
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

func (pool *LinkPool) subscribe(match func(*Stream, *string) bool) *streamWatcher {
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

func (pool *LinkPool) notifyStreamWatchers(s *Stream, strategy *string) bool {
	used := false
	for _, w := range pool.watchers.Clone() {
		if !w.match(s, strategy) {
			continue
		}

		select {
		case w.ch <- s:
			used = true
		default:
		}
	}

	return used
}

func (pool *LinkPool) getOrCreateNodeLinker(target *astral.Identity) *NodeLinker {
	key := target.String()

	if linker, ok := pool.linkers.Get(key); ok {
		return linker
	}

	linker := NewNodeLinker(pool.mod, target)
	if existing, ok := pool.linkers.Set(key, linker); !ok {
		return existing
	}

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

	match := func(s *Stream, strategy *string) bool {
		if !s.RemoteIdentity().IsEqual(target) {
			return false
		}
		if strategy != nil && len(o.Strategies) > 0 {
			return slices.Contains(o.Strategies, *strategy)
		}

		if len(o.Networks) > 0 {
			return slices.Contains(o.Networks, s.Network())
		}
		return true
	}

	if !o.ForceNew {
		streams := pool.peers.streams.Select(func(s *Stream) bool { return match(s, nil) })
		if len(streams) > 0 {
			return sig.ArrayToChan([]LinkResult{{Stream: streams[0]}})
		}
	}

	result := make(chan LinkResult, 1)

	go func() {
		defer close(result)

		strategyCtx, cancelStrategies := ctx.WithCancel()
		defer cancelStrategies()

		w := pool.subscribe(match)
		defer pool.unsubscribe(w)

		linker := pool.getOrCreateNodeLinker(target)
		done := linker.Activate(strategyCtx, o.Strategies)

		select {
		case <-done:
			select {
			case s := <-w.ch:
				result <- LinkResult{Stream: s}
			default:
				result <- LinkResult{Err: nodes.ErrStreamNotProduced}
			}
		case <-ctx.Done():
			result <- LinkResult{Err: ctx.Err()}
		case s := <-w.ch:
			result <- LinkResult{Stream: s}
		}
	}()

	return result
}

// RetrieveLinkOptions controls how RetrieveLink behaves.
type RetrieveLinkOptions struct {
	ForceNew   bool
	Strategies []string
	Networks   []string
}

// RetrieveLinkOption is a functional option for RetrieveLink.
type RetrieveLinkOption func(*RetrieveLinkOptions)

func WithForceNew() RetrieveLinkOption {
	return func(o *RetrieveLinkOptions) {
		o.ForceNew = true
	}
}

func WithStrategies(strategies ...string) RetrieveLinkOption {
	return func(o *RetrieveLinkOptions) {
		o.Strategies = append(o.Strategies, strategies...)
	}
}

func WithNetworks(networks ...string) RetrieveLinkOption {
	return func(o *RetrieveLinkOptions) {
		o.Networks = append(o.Networks, networks...)
	}
}
